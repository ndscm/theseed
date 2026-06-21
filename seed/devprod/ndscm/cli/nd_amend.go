package main

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndAmendFlags struct {
	breakText    *seedflag.StringFlag
	migrateText  *seedflag.StringFlag
	sideEffectOf *seedflag.StringFlag
}

func parseNdAmendFlags(args []string) (ndAmendFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-amend")
	cmdFlags := ndAmendFlags{}
	cmdFlags.breakText = cf.DefineString("break", "", "Breaking change description")
	cmdFlags.migrateText = cf.DefineString("migrate", "", "Migration instruction")
	cmdFlags.sideEffectOf = cf.DefineString("side-effect-of", "", "Change UUID this commit is a side effect of")
	cmdArgs, err := cf.Parse(args)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndAmend(args []string) error {
	cmdFlags, cmdArgs, err := parseNdAmendFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 0 {
		return seederr.WrapErrorf("nd-amend usage: nd amend --side-effect-of=<uuid>|break --break=<text> --migrate=<text>")
	}
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdAmend(clientcore.NdAmendOptions{
		Break:        cmdFlags.breakText.Get(),
		Migrate:      cmdFlags.migrateText.Get(),
		SideEffectOf: cmdFlags.sideEffectOf.Get(),
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
