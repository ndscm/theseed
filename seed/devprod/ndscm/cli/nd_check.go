package main

import (
	"runtime"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndCheckFlags struct {
	all *seedflag.BoolFlag
}

func parseNdCheckFlags(args []string) (ndCheckFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-check")
	cmdFlags := ndCheckFlags{}
	cmdFlags.all = cf.DefineBool("all", false, "Check all watchers regardless of cached modified time")
	cmdArgs, err := cf.Parse(args,
		seedflag.WithAnywhereFlag(true),
	)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndCheck(args []string) error {
	cmdFlags, cmdArgs, err := parseNdCheckFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 0 {
		return seederr.WrapErrorf("nd-check usage: ndscm check")
	}
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}

	err = cc.NdRun(clientcore.NdRunOptions{
		Workers: runtime.NumCPU(),
		All:     cmdFlags.all.Get(),
		Changed: true,

		Phases: []string{"format"},
	})
	if err != nil {
		return seederr.Wrap(err)
	}

	err = cc.NdRun(clientcore.NdRunOptions{
		Workers: runtime.NumCPU(),
		All:     cmdFlags.all.Get(),
		Changed: false,

		Phases: []string{"vendor", "bootstrap", "tidy", "lock", "build", "test"},
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
