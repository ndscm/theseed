package ageflag

import (
	"bytes"
	"io"

	"filippo.io/age"
	"filippo.io/age/armor"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type AgeSecretFlag struct {
	stringFlag *seedflag.StringFlag
	fileFlag   *seedflag.FileFlag
}

func (f *AgeSecretFlag) Set(s string) error {
	return f.stringFlag.Set(s)
}

func (f *AgeSecretFlag) Get() (string, string) {
	return f.stringFlag.Get(), f.fileFlag.Get()
}

func (f *AgeSecretFlag) String() string {
	return "**REDACTED** " + f.fileFlag.String()
}

func (f *AgeSecretFlag) Load() ([]byte, error) {
	value := f.stringFlag.Get()
	if value != "" {
		return []byte(value), nil
	}
	scheme, fileBytes, err := f.fileFlag.LoadScheme()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if scheme == "age" {
		identities, err := LoadIdentities()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		sourceReader := (io.Reader)(bytes.NewReader(fileBytes))
		if bytes.HasPrefix(bytes.TrimSpace(fileBytes), []byte(armor.Header)) {
			sourceReader = armor.NewReader(sourceReader)
		}
		plainReader, err := age.Decrypt(sourceReader, identities...)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		fileBytes, err = io.ReadAll(plainReader)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
	}
	return fileBytes, nil
}

func (f *AgeSecretFlag) LoadString() (string, error) {
	fileBytes, err := f.Load()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return string(fileBytes), nil
}

func NewAgeSecretFlag(usage string) *AgeSecretFlag {
	stringFlag := seedflag.NewStringFlag("", usage)
	fileFlag := seedflag.NewFileFlag("", usage)
	fileFlag.RegisterLocalScheme("age")
	return &AgeSecretFlag{
		stringFlag: stringFlag,
		fileFlag:   fileFlag,
	}
}

func DefineSecret(name string, usage string) *AgeSecretFlag {
	item := NewAgeSecretFlag(usage)
	// TODO(nagi): support susceptible secret flag for loading secret from cli or env.
	seedflag.DefineFlag(name+"_file", item.fileFlag)
	return item
}
