package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/groovy"
	"github.com/ndscm/theseed/seed/infra/groovy/generator/gengroovy"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagWrite = seedflag.DefineBool("write", false, "Write changes to files")

func run() error {
	args, err := seedinit.Initialize(
		seedinit.WithAnywhereFlag(true),
	)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(args) < 1 {
		return seederr.WrapErrorf("groovy-format usage: groovy-format file")
	}
	for _, path := range args {
		source, err := os.ReadFile(path)
		if err != nil {
			return seederr.Wrap(err)
		}
		module, err := groovy.Parse(string(source))
		if err != nil {
			return seederr.Wrap(err)
		}
		formatted, err := gengroovy.Generate(module)
		if err != nil {
			return seederr.Wrap(err)
		}
		formatted = strings.TrimRight(formatted, "\n")
		if formatted != "" {
			formatted += "\n"
		}
		if flagWrite.Get() {
			err := os.WriteFile(path, []byte(formatted), 0644)
			if err != nil {
				return seederr.Wrap(err)
			}
		} else {
			_, err = fmt.Print(formatted)
			if err != nil {
				return seederr.Wrap(err)
			}
		}
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		seedlog.Errorf("%v", err)
		os.Exit(1)
	}
}
