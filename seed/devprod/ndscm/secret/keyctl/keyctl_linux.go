//go:build linux

package keyctl

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/secret"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
	"golang.org/x/sys/unix"
)

type KeyctlProvider struct {
}

func (k *KeyctlProvider) searchNdscmKeyring() (string, error) {
	ndscmKeyringIdBytes, err := seedshell.PureOutput("keyctl", "search", "@s", "keyring", "ndscm")
	if err != nil {
		// keyctl exits 1 when the requested key is not available (ENOKEY); treat
		// that as "no ndscm keyring yet" and report it as not found rather than a
		// hard error. Other exit codes (e.g. 2 for malformed args) are real
		// failures and propagate.
		exitErr := (*exec.ExitError)(nil)
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", nil
		}
		return "", seederr.Wrap(err)
	}
	ndscmKeyringId := strings.TrimSpace(string(ndscmKeyringIdBytes))
	return ndscmKeyringId, nil
}

func (k *KeyctlProvider) forkSession() error {
	ndscmKeyringId, err := k.searchNdscmKeyring()
	if err != nil {
		return seederr.Wrap(err)
	}
	if ndscmKeyringId == "" {
		// Fork the parent shell's current session keyring into a fresh session,
		// then push it onto the parent with KEYCTL_SESSION_TO_PARENT.
		//
		// A plain `keyctl new_session` (JOIN_SESSION_KEYRING + SESSION_TO_PARENT)
		// installs an *empty* keyring onto the shell, dropping every key the old
		// session held and breaking other programs that depend on them. We want
		// the new session to carry those keys over.
		//
		// The catch is keyring possession. A key's permissions are evaluated
		// against the possessor bits only while the key is reachable from one of
		// our keyrings; otherwise the (far more restrictive) owner bits apply,
		// which for typical keys grant neither READ nor LINK. JOIN_SESSION_KEYRING
		// swaps out our session keyring, so the instant it returns we no longer
		// possess the old keys and can neither read nor re-link them.
		//
		// JOIN_SESSION_KEYRING only replaces the *session* keyring, though -- it
		// leaves the process keyring untouched. So we stash the old session
		// keyring in the process keyring first, keeping possession of it (and
		// transitively all its keys) across the join, then nest it into the fresh
		// session with a single link and drop the temporary stash.
		oldSessionKeyringId, err := unix.KeyctlGetKeyringID(unix.KEY_SPEC_SESSION_KEYRING, false)
		if err != nil {
			return seederr.Wrap(err)
		}

		procKeyringId, err := unix.KeyctlGetKeyringID(unix.KEY_SPEC_PROCESS_KEYRING, true)
		if err != nil {
			return seederr.Wrap(err)
		}
		_, err = unix.KeyctlInt(unix.KEYCTL_LINK, oldSessionKeyringId, procKeyringId, 0, 0)
		if err != nil {
			return seederr.Wrap(err)
		}

		// Mirror the original session's name onto the fork instead of hardcoding
		// "_ses". KEYCTL_DESCRIBE returns "type;uid;gid;perm;description"; the
		// keyring's name is the final ';'-separated field.
		oldSessionDescribe, err := unix.KeyctlString(unix.KEYCTL_DESCRIBE, oldSessionKeyringId)
		if err != nil {
			return seederr.Wrap(err)
		}
		_, oldSessionName, _ := strings.CutLast(oldSessionDescribe, ";")

		// Join a fresh session. We still possess the old session keyring via the
		// process keyring, so nesting it below succeeds under possessor perms.
		// The old keys remain reachable (keyctl search recurses into subkeyrings).
		newSessionKeyringId, err := unix.KeyctlJoinSessionKeyring(oldSessionName)
		if err != nil {
			return seederr.Wrap(err)
		}
		_, err = unix.KeyctlInt(unix.KEYCTL_LINK, oldSessionKeyringId, newSessionKeyringId, 0, 0)
		if err != nil {
			return seederr.Wrap(err)
		}

		// The old session is now possessed via the new session; the process-keyring
		// stash has done its job, so unlink it to leave no extra references.
		_, _ = unix.KeyctlInt(unix.KEYCTL_UNLINK, oldSessionKeyringId, procKeyringId, 0, 0)

		// Push the forked session onto the parent (the invoking shell), so every
		// command it later spawns inherits it.
		_, _, errno := unix.Syscall(unix.SYS_KEYCTL, uintptr(unix.KEYCTL_SESSION_TO_PARENT), 0, 0)
		if errno != 0 {
			return seederr.Wrap(errno)
		}
		seedlog.Infof("Joined keyctl session: %v", newSessionKeyringId)

		err = seedshell.ImpureRun("keyctl", "newring", "ndscm", "@s")
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	ndscmKeyringId, err = k.searchNdscmKeyring()
	if err != nil {
		return seederr.Wrap(err)
	}
	if ndscmKeyringId == "" {
		return seederr.WrapErrorf("failed to initialize keyctl session")
	}
	return nil
}

func (k *KeyctlProvider) Save(key string, value []byte) error {
	ndscmKeyringId, err := k.searchNdscmKeyring()
	if err != nil {
		return seederr.Wrap(err)
	}
	if ndscmKeyringId == "" {
		return seederr.WrapErrorf("keyctl session not initialized")
	}

	stdinOption := func(cmd *exec.Cmd) {
		cmd.Stdin = bytes.NewReader(value)
	}
	err = seedshell.ImpureOptionsRun(
		[]seedshell.RunOption{stdinOption},
		"keyctl", "padd", "user", key, ndscmKeyringId,
	)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func (k *KeyctlProvider) Load(key string) ([]byte, error) {
	ndscmKeyringId, err := k.searchNdscmKeyring()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if ndscmKeyringId == "" {
		return nil, seederr.WrapErrorf("keyctl session not initialized")
	}

	keyIdBytes, err := seedshell.PureOutput("keyctl", "search", ndscmKeyringId, "user", key)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	keyId := strings.TrimSpace(string(keyIdBytes))
	valueBytes, err := seedshell.PureOutput("keyctl", "pipe", keyId)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return valueBytes, nil
}

var _ secret.SecretSessionProvider = (*KeyctlProvider)(nil)

func initialize() (secret.SecretSessionProvider, error) {
	_, err := exec.LookPath("keyctl")
	if err != nil {
		// Skipped if keyctl is not installed.
		return nil, nil
	}

	keyctlProvider := &KeyctlProvider{}
	err = keyctlProvider.forkSession()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return keyctlProvider, nil
}

func init() {
	secret.RegisterSessionProvider("keyctl", initialize)
}
