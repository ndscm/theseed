package main

import (
	"fmt"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

func ndTestable(args []string) error {
	seedflag.Finalize(args)
	commit := ""
	if len(args) > 1 {
		return seederr.WrapErrorf("nd-testable usage: ndscm testable [commit]")
	}
	if len(args) == 1 {
		commit = args[0]
	}

	cc := &clientcore.ClientCore{}
	err := cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}

	testable, err := cc.CheckTestable(commit)
	if err != nil {
		return seederr.Wrap(err)
	}
	fmt.Printf("%v\n", testable)
	return nil
}
