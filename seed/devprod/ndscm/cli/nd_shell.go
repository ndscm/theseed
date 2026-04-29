package main

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

func ndShell(args []string) error {
	seedflag.Finalize(args)
	if len(args) != 0 {
		return seederr.WrapErrorf("nd-shell usage: eval \"$(ndscm --shell-eval shell)\"")
	}
	cc := &clientcore.ClientCore{}
	err := cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdShell(clientcore.NdShellOptions{})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
