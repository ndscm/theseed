package main

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndSyncFlags struct {
	fetch *seedflag.StringFlag
}

func parseNdSyncFlags(args []string) (ndSyncFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-sync")
	cmdFlags := ndSyncFlags{}
	cmdFlags.fetch = cf.DefineString("fetch", "auto", "When to fetch upstream changes before rebasing: always, never, or auto (fetch only on the dev branch tip)")
	cmdArgs, err := cf.Parse(args,
		seedflag.WithAnywhereFlag(true),
	)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndSync(args []string) error {
	cmdFlags, cmdArgs, err := parseNdSyncFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 0 {
		return seederr.WrapErrorf("nd-sync usage: (on dev branch) nd sync")
	}
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdSync(clientcore.NdSyncOptions{
		Fetch: cmdFlags.fetch.Get(),
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
