package main

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

func ndSync(args []string) error {
	seedflag.Finalize(args)
	if len(args) != 0 {
		return seederr.WrapErrorf("nd-sync usage: (on dev branch) nd sync")
	}
	cc := &clientcore.ClientCore{}
	err := cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdSync(clientcore.NdSyncOptions{})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
