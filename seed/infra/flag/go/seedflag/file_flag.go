package seedflag

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type FileFlag struct {
	StringFlag
}

func (f *FileFlag) Set(s string) error {
	f.value = s
	return nil
}

func (f *FileFlag) Get() string {
	return f.value
}

func (f *FileFlag) String() string {
	return f.value
}

func (f *FileFlag) Load() ([]byte, error) {
	filePath := strings.TrimSpace(f.value)
	if filePath == "" {
		return nil, nil
	}
	if strings.HasPrefix(filePath, "~/") {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		filePath = filepath.Join(userHomeDir, filePath[2:])
	}
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return fileBytes, nil
}

func (f *FileFlag) LoadString() (string, error) {
	fileBytes, err := f.Load()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(fileBytes)), nil
}

var _ FlagDefinition = (*FileFlag)(nil)

func newFileFlag(defaultValue string, usage string) *FileFlag {
	return &FileFlag{
		StringFlag: StringFlag{
			FlagItem: FlagItem{
				usage: usage,
			},
			value: defaultValue,
		},
	}
}
