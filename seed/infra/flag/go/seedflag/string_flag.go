package seedflag

type StringFlag struct {
	FlagItem
	value string
}

func (f *StringFlag) Set(s string) error {
	f.value = s
	return nil
}

func (f *StringFlag) Get() string {
	return f.value
}

func (f *StringFlag) String() string {
	return f.value
}

var _ FlagDefinition = (*StringFlag)(nil)

func NewStringFlag(defaultValue string, usage string) *StringFlag {
	return &StringFlag{
		FlagItem: FlagItem{
			usage: usage,
		},
		value: defaultValue,
	}
}
