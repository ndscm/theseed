package sopsflag

import (
	"github.com/getsops/sops/v3/decrypt"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type SopsSecretFlag struct {
	stringFlag *seedflag.StringFlag
	fileFlag   *seedflag.FileFlag
}

func (f *SopsSecretFlag) Set(s string) error {
	return f.stringFlag.Set(s)
}

func (f *SopsSecretFlag) Get() (string, string) {
	return f.stringFlag.Get(), f.fileFlag.Get()
}

func (f *SopsSecretFlag) String() string {
	return "**REDACTED**" + f.fileFlag.String()
}

func (f *SopsSecretFlag) Load() ([]byte, error) {
	value := f.stringFlag.Get()
	if value != "" {
		return []byte(value), nil
	}
	scheme, fileBytes, err := f.fileFlag.LoadScheme()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if scheme == "sops" {
		fileBytes, err = decrypt.Data(fileBytes, "binary")
		if err != nil {
			return nil, seederr.Wrap(err)
		}
	}
	return fileBytes, nil
}

func (f *SopsSecretFlag) LoadString() (string, error) {
	fileBytes, err := f.Load()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return string(fileBytes), nil
}

func NewSopsSecretFlag(usage string) *SopsSecretFlag {
	stringFlag := seedflag.NewStringFlag("", usage)
	fileFlag := seedflag.NewFileFlag("", usage)
	fileFlag.RegisterLocalScheme("sops")
	return &SopsSecretFlag{
		stringFlag: stringFlag,
		fileFlag:   fileFlag,
	}
}

func DefineSecret(name string, usage string) *SopsSecretFlag {
	item := NewSopsSecretFlag(usage)
	// TODO(nagi): support susceptible secret flag for loading secret from cli or env.
	seedflag.DefineFlag(name+"_file", item.fileFlag)
	return item
}
