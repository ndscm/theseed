package seedflag

import (
	"flag"
)

type FlagDefinition interface {
	flag.Value

	Usage() string
}

type FlagItem struct {
	usage string
}

func (f *FlagItem) Usage() string {
	return f.usage
}
