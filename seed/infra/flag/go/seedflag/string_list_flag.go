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
	// Each value is split on commas so a whole list can be supplied by a single
	// occurrence — e.g. an env var FOO=a,b,c or one --foo=a,b,c on the command
	// line — in addition to repeating the flag. Entries are trimmed and empty
	// ones (from surrounding, trailing, or doubled commas) are dropped.
	//
	// Because commas are always separators here, a list element cannot itself
	// contain a comma. If an individual value must carry commas, use a
	// StringFlag or FileFlag instead of a StringListFlag.
	s = strings.TrimSpace(s)
	if s != "" {
		parts := strings.Split(s, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				f.values = append(f.values, part)
			}
		}
	}
	return nil
}

func (f *StringListFlag) Get() []string {
	return f.values
}

func (f *StringListFlag) String() string {
	return "[" + strings.Join(f.values, ", ") + "]"
}

var _ FlagDefinition = (*StringListFlag)(nil)

func NewStringListFlag(defaultValues []string, usage string) *StringListFlag {
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
