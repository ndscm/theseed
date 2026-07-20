package main

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndSecretFlags struct {
	space *seedflag.StringFlag
	user  *seedflag.BoolFlag
}

func parseNdSecretFlags(args []string) (ndSecretFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-secret")
	cmdFlags := ndSecretFlags{}
	cmdFlags.space = cf.DefineString("space", "main", "The secret space; the worktree is secret/<space>")
	cmdFlags.user = cf.DefineBool("user", false, "Scope the secret worktree to the current user: <user-handle>/secret/<space>")
	cmdArgs, err := cf.Parse(args,
		seedflag.WithAnywhereFlag(true),
	)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndSecret(args []string) error {
	cmdFlags, cmdArgs, err := parseNdSecretFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdSecret(clientcore.NdSecretOptions{
		Space: cmdFlags.space.Get(),
		User:  cmdFlags.user.Get(),

		Args: cmdArgs,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
