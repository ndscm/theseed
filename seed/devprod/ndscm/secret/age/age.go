package age

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/secret"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
	"golang.org/x/term"
)

// `age-keygen -pq | age -p -a > key.age`
func generateKey(keyPath string) error {
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer keyFile.Close()

	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		return seederr.Wrap(err)
	}
	defer pipeReader.Close()
	keygenCmd := exec.Command("age-keygen", "-pq")
	keygenCmd.Stdout = pipeWriter
	keygenCmd.Stderr = os.Stderr
	encryptCmd := exec.Command("age", "-p", "-a")
	encryptCmd.Stdin = pipeReader
	encryptCmd.Stdout = keyFile
	encryptCmd.Stderr = os.Stderr

	err = encryptCmd.Start()
	if err != nil {
		pipeWriter.Close()
		return seederr.Wrap(err)
	}

	keygenErr := keygenCmd.Run()
	// Close the write end so the encrypt child sees EOF. This must precede
	// encryptCmd.Wait() below, so it cannot be deferred.
	pipeWriter.Close()

	encryptErr := encryptCmd.Wait()
	if keygenErr != nil {
		return seederr.Wrap(keygenErr)
	}
	if encryptErr != nil {
		return seederr.Wrap(encryptErr)
	}
	return nil
}

// `age -d key.age | age-keygen -y`
func populateRecipient(keyPath string) (string, error) {
	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	defer pipeReader.Close()
	var recipient bytes.Buffer
	decryptCmd := exec.Command("age", "-d", keyPath)
	decryptCmd.Stdout = pipeWriter
	decryptCmd.Stderr = os.Stderr
	pubkeyCmd := exec.Command("age-keygen", "-y")
	pubkeyCmd.Stdin = pipeReader
	pubkeyCmd.Stdout = &recipient
	pubkeyCmd.Stderr = os.Stderr

	err = pubkeyCmd.Start()
	if err != nil {
		pipeWriter.Close()
		return "", seederr.Wrap(err)
	}

	decryptErr := decryptCmd.Run()
	// Close the write end so the pubkey child sees EOF. This must precede
	// pubkeyCmd.Wait() below, so it cannot be deferred.
	pipeWriter.Close()

	pubkeyErr := pubkeyCmd.Wait()
	if decryptErr != nil {
		return "", seederr.Wrap(decryptErr)
	}
	if pubkeyErr != nil {
		return "", seederr.Wrap(pubkeyErr)
	}
	return strings.TrimSpace(recipient.String()), nil
}

// appendRecipient appends recipient as a new line to the recipients file,
// creating it if necessary so any existing recipients are preserved. If the
// existing file does not end in a newline, a separating newline is written
// first so the new recipient starts on its own line.
func appendRecipient(recipientsPath string, recipient string) error {
	exists := false
	info, err := os.Stat(recipientsPath)
	if err == nil {
		exists = true
		seedlog.Warnf("recipients file already exists, appending: %v", recipientsPath)
	} else if !os.IsNotExist(err) {
		return seederr.Wrap(err)
	}

	recipientsFile := (*os.File)(nil)
	if exists {
		recipientsFile, err = os.OpenFile(recipientsPath, os.O_RDWR|os.O_APPEND, 0o644)
		if err != nil {
			return seederr.Wrap(err)
		}
		defer recipientsFile.Close()
		// A non-empty file whose final byte is not a newline needs a separating
		// newline before the appended recipient. ReadAt uses an absolute offset,
		// so it does not disturb the append position.
		if size := info.Size(); size > 0 {
			lastByte := make([]byte, 1)
			_, err = recipientsFile.ReadAt(lastByte, size-1)
			if err != nil {
				return seederr.Wrap(err)
			}
			if lastByte[0] != '\n' {
				_, err = recipientsFile.WriteString("\n")
				if err != nil {
					return seederr.Wrap(err)
				}
			}
		}
	} else {
		recipientsFile, err = os.OpenFile(recipientsPath, os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return seederr.Wrap(err)
		}
		defer recipientsFile.Close()
	}

	_, err = recipientsFile.WriteString(recipient + "\n")
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

type AgeProvider struct{}

func (a *AgeProvider) Keygen(worktreePath string) error {
	keyPath := filepath.Join(worktreePath, "key.age")
	recipientsPath := filepath.Join(worktreePath, "recipients.txt")

	_, err := os.Stat(keyPath)
	if err == nil {
		return seederr.WrapErrorf("age key already exists at %v", keyPath)
	}
	if !os.IsNotExist(err) {
		return seederr.Wrap(err)
	}
	if seedshell.Dry() {
		return nil
	}

	err = generateKey(keyPath)
	if err != nil {
		return seederr.Wrap(err)
	}
	recipient, err := populateRecipient(keyPath)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = appendRecipient(recipientsPath, recipient)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = seedshell.PureRun("age-inspect", "--json", keyPath)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func (a *AgeProvider) Encrypt(worktreePath string, secretPath string) error {
	recipientsAbsPath := filepath.Join(worktreePath, "recipients.txt")
	secretAbsPath := filepath.Join(worktreePath, secretPath)
	if seedshell.Dry() {
		seedlog.Infof("Dry mode skip: encrypt %v", secretAbsPath)
		return nil
	}
	err := os.MkdirAll(filepath.Dir(secretAbsPath), 0o755)
	if err != nil {
		return seederr.Wrap(err)
	}

	// When stdin is a terminal, disable echoing while age reads the secret (like
	// sudo) so it never appears on screen. The plaintext still streams straight
	// from stdin into age and never lands in a Go buffer; input ends at EOF
	// (Ctrl-D). When stdin is piped there is nothing to echo.
	stdinFd := os.Stdin.Fd()
	if term.IsTerminal(int(stdinFd)) {
		_, err := fmt.Fprint(os.Stderr, "Enter secret (Use Ctrl-D twice to end):")
		if err != nil {
			return seederr.Wrap(err)
		}
		restore, err := disableEcho(stdinFd)
		if err != nil {
			return seederr.Wrap(err)
		}
		defer restore()
	}
	stdinOption := func(cmd *exec.Cmd) {
		cmd.Stdin = os.Stdin
	}
	err = seedshell.ImpureOptionsRun(
		[]seedshell.RunOption{stdinOption},
		"age", "-e", "-R", recipientsAbsPath, "-o", secretAbsPath,
	)
	if err != nil {
		return seederr.Wrap(err)
	}
	_, err = fmt.Fprintln(os.Stderr, "")
	if err != nil {
		return seederr.Wrap(err)
	}
	err = seedshell.PureRun("age-inspect", "--json", secretAbsPath)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func (a *AgeProvider) Decrypt(worktreePath string, secretPath string) error {
	ageKeyFile := os.Getenv("AGE_KEY_FILE")
	if ageKeyFile == "" {
		ageKeyFile = filepath.Join(worktreePath, "key.age")
	}
	secretAbsPath := filepath.Join(worktreePath, secretPath)
	err := seedshell.PureRun("age", "-d", "-i", ageKeyFile, secretAbsPath)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

var _ secret.Provider = (*AgeProvider)(nil)
