package seedflag

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
	value := f.stringFlag.Get()
	if value != "" {
		return value, nil
	}
	return f.fileFlag.LoadString()
}

func newSecretFlag(usage string) *SecretFlag {
	return &SecretFlag{
		stringFlag: newStringFlag("", usage),
		fileFlag:   newFileFlag("", usage),
	}
}
