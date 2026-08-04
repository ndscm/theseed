package main

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndTicketFlags struct {
	space *seedflag.StringFlag
}

func parseNdTicketFlags(args []string) (ndTicketFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-ticket")
	cmdFlags := ndTicketFlags{}
	cmdFlags.space = cf.DefineString("space", "main", "The ticket space; the worktree is ticket/<space>")
	cmdArgs, err := cf.Parse(args,
		seedflag.WithAnywhereFlag(true),
	)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndTicket(args []string) error {
	cmdFlags, cmdArgs, err := parseNdTicketFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdTicket(clientcore.NdTicketOptions{
		Space: cmdFlags.space.Get(),

		Args: cmdArgs,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
