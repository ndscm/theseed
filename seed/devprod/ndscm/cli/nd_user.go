package main

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

func ndUser(args []string) error {
	cc := &clientcore.ClientCore{}
	err := cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdUser(clientcore.NdUserOptions{
		Args: args,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
