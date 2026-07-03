package seedflag

import "strings"

type StringListFlag struct {
	FlagItem
	values []string
	set    bool
}

func (f *StringListFlag) Set(s string) error {
	// The first explicit Set replaces the defaults rather than appending to
	// them, so a supplied value fully overrides the default list. Subsequent
	// Sets (repeated flag occurrences) accumulate.
	if !f.set {
		f.values = nil
		f.set = true
	}
	f.values = append(f.values, s)
	return nil
}

func (f *StringListFlag) Get() []string {
	return f.values
}

func (f *StringListFlag) String() string {
	return "[" + strings.Join(f.values, ", ") + "]"
}

var _ FlagDefinition = (*StringListFlag)(nil)

func newStringListFlag(defaultValues []string, usage string) *StringListFlag {
	values := defaultValues
	if values == nil {
		values = []string{}
	}
	return &StringListFlag{
		FlagItem: FlagItem{
			usage: usage,
		},
		values: values,
	}
}
