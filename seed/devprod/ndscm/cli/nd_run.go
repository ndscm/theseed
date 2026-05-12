package main

import (
	"runtime"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndRunFlags struct {
	all     *seedflag.BoolFlag
	changed *seedflag.BoolFlag
}

func parseNdRunFlags(args []string) (ndRunFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-run")
	cmdFlags := ndRunFlags{}
	cmdFlags.all = cf.DefineBool("all", false, "Run against all files instead of only dirty files")
	cmdFlags.changed = cf.DefineBool("changed", false, "Run against changed files instead of only dirty files")
	cmdArgs, err := cf.Parse(args,
		seedflag.WithAnywhereFlag(true),
	)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndRun(args []string) error {
	cmdFlags, cmdArgs, err := parseNdRunFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) == 0 {
		return seederr.WrapErrorf("nd-run usage: ndscm run <phase> [<phase>...]")
	}
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdRun(clientcore.NdRunOptions{
		Workers: runtime.NumCPU(),
		All:     cmdFlags.all.Get(),
		Changed: cmdFlags.changed.Get(),

		Phases: cmdArgs,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
