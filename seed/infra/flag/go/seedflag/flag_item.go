package seedflag

import (
	"flag"
)

type FlagDefinition interface {
	Usage() string
	flag.Value
}

type FlagItem struct {
	usage string
}

func (f *FlagItem) Usage() string {
	return f.usage
}
