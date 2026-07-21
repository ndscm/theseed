package seedflag

import (
	"strconv"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type BoolFlag struct {
	FlagItem
	value bool
}

func (f *BoolFlag) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return seederr.Wrap(err)
	}
	f.value = v
	return nil
}

func (f *BoolFlag) Get() bool {
	return f.value
}

func (f *BoolFlag) String() string {
	return strconv.FormatBool(f.value)
}

func (f *BoolFlag) IsBoolFlag() bool {
	return true
}

var _ FlagDefinition = (*BoolFlag)(nil)

func NewBoolFlag(defaultValue bool, usage string) *BoolFlag {
	return &BoolFlag{
		FlagItem: FlagItem{
			usage: usage,
		},
		value: defaultValue,
	}
}
