package seedflag

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

// FileFlag holds a path to a file whose contents are read on demand via Load.
// The value is parsed as a URI. Its scheme selects how the caller interprets
// the bytes; FileFlag itself always reads the file as-is and reports the scheme
// back (via LoadScheme), leaving any further handling to the caller.
//
// Only registered schemes are accepted; any other scheme is rejected. By
// default "" (a plain path) and "file" are registered; additional schemes are
// added with RegisterLocalScheme. For example sopsflag registers "sops" and
// SOPS-decrypts the bytes LoadScheme returns.
//
// A leading "~/" is expanded to the user's home directory.
//
// Because the value is parsed with net/url, the path is interpreted more
// permissively than the filesystem would: percent escapes are decoded (so
// "my%20file" reads "my file"), a "?" or "#" begins a query or fragment and is
// dropped from the path, and a leading "//" is read as a host and lost. Typical
// config paths ("/etc/...", "~/...", "./rel") contain none of these, but a path
// that does must be file-scheme encoded (or avoided).
type FileFlag struct {
	StringFlag

	localSchemes map[string]bool
}

// Set records the raw flag value (a path or file URI). It implements flag.Value
// and does not touch the filesystem; the file is read later by Load.
func (f *FileFlag) Set(s string) error {
	f.value = s
	return nil
}

// Get returns the raw flag value as set, before any file is read.
func (f *FileFlag) Get() string {
	return f.value
}

// String returns the raw flag value; it implements flag.Value.
func (f *FileFlag) String() string {
	return f.value
}

// RegisterLocalScheme adds scheme to the set of URI schemes the flag accepts,
// beyond the "" and "file" schemes registered by default. The file is still
// read as-is; the scheme is only reported back by LoadScheme for the caller to
// act on.
func (f *FileFlag) RegisterLocalScheme(scheme string) {
	f.localSchemes[scheme] = true
}

// LoadScheme reads the file the flag points at and returns its URI scheme
// alongside the contents. An empty value yields an empty scheme, a nil slice,
// and no error. An unregistered scheme is an error. See FileFlag for how the
// value is resolved to a path and the caveats that apply to it.
func (f *FileFlag) LoadScheme() (string, []byte, error) {
	raw := strings.TrimSpace(f.value)
	if raw == "" {
		return "", nil, nil
	}
	fileUrl, err := url.Parse(raw)
	if err != nil {
		return "", nil, seederr.Wrap(err)
	}

	if !f.localSchemes[fileUrl.Scheme] {
		return "", nil, seederr.WrapErrorf("unsupported file flag scheme %q: %s", fileUrl.Scheme, raw)
	}
	localPath := fileUrl.Path

	if strings.HasPrefix(localPath, "~/") {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return "", nil, seederr.Wrap(err)
		}
		localPath = filepath.Join(userHomeDir, localPath[2:])
	}
	fileBytes, err := os.ReadFile(localPath)
	if err != nil {
		return "", nil, seederr.Wrap(err)
	}
	return fileUrl.Scheme, fileBytes, nil
}

// Load reads and returns the contents of the file the flag points at,
// discarding the scheme. An empty value yields a nil slice and no error. See
// FileFlag for how the value is resolved to a path and the caveats that apply.
func (f *FileFlag) Load() ([]byte, error) {
	_, fileBytes, err := f.LoadScheme()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return fileBytes, nil
}

// LoadSchemeString is like LoadScheme but returns the file contents as a string
// with leading and trailing whitespace trimmed. An empty value yields an empty
// scheme and "".
func (f *FileFlag) LoadSchemeString() (string, string, error) {
	scheme, fileBytes, err := f.LoadScheme()
	if err != nil {
		return "", "", seederr.Wrap(err)
	}
	return scheme, strings.TrimSpace(string(fileBytes)), nil
}

// LoadString is like Load but returns the file contents as a string with
// leading and trailing whitespace trimmed. An empty value yields "".
func (f *FileFlag) LoadString() (string, error) {
	_, fileString, err := f.LoadSchemeString()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return fileString, nil
}

var _ FlagDefinition = (*FileFlag)(nil)

// NewFileFlag returns a FileFlag with the given default value and usage string.
// Only the "" and "file" schemes are accepted; use RegisterLocalScheme to add
// more.
func NewFileFlag(defaultValue string, usage string) *FileFlag {
	return &FileFlag{
		StringFlag: StringFlag{
			FlagItem: FlagItem{
				usage: usage,
			},
			value: defaultValue,
		},
		localSchemes: map[string]bool{
			"":     true,
			"file": true,
		},
	}
}
