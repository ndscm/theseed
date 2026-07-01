package main

import (
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

func ndCommit(args []string) error {
	seedflag.Finalize(args)
	if len(args) == 0 {
		return seederr.WrapErrorf("nd-commit usage: nd commit <message>")
	}
	message := strings.Join(args, " ")
	cc := &clientcore.ClientCore{}
	err := cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdCommit(clientcore.NdCommitOptions{
		Message: message,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
