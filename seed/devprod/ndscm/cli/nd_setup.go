package main

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

func ndSetup(args []string) error {
	seedflag.Finalize(args)
	if len(args) != 0 {
		return seederr.WrapErrorf("nd-setup usage: ndscm setup")
	}
	cc := &clientcore.ClientCore{}
	err := cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdSetup(clientcore.NdSetupOptions{})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
