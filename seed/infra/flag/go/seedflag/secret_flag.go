package seedflag

import (
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type SecretFlag struct {
	stringFlag *StringFlag
	fileFlag   *FileFlag
}

func (f *SecretFlag) Set(s string) error {
	return f.stringFlag.Set(s)
}

func (f *SecretFlag) Get() (string, string) {
	return f.stringFlag.Get(), f.fileFlag.Get()
}

func (f *SecretFlag) String() string {
	return "**REDACTED**" + f.fileFlag.String()
}

func (f *SecretFlag) Load() ([]byte, error) {
	value := f.stringFlag.Get()
	if value != "" {
		return []byte(value), nil
	}
	return f.fileFlag.Load()
}

func (f *SecretFlag) LoadString() (string, error) {
	fileBytes, err := f.Load()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return string(fileBytes), nil
}

func NewSecretFlag(usage string) *SecretFlag {
	return &SecretFlag{
		stringFlag: NewStringFlag("", usage),
		fileFlag:   NewFileFlag("", usage),
	}
}
