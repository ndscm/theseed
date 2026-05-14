package main

import (
	"runtime"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndLockFlags struct {
	all *seedflag.BoolFlag
}

func parseNdLockFlags(args []string) (ndLockFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-lock")
	cmdFlags := ndLockFlags{}
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

func ndLock(args []string) error {
	cmdFlags, cmdArgs, err := parseNdLockFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 0 {
		return seederr.WrapErrorf("nd-lock usage: ndscm lock")
	}
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdRun(clientcore.NdRunOptions{
		Workers: runtime.NumCPU(),
		All:     cmdFlags.all.Get(),
		Changed: false,

		Phases: []string{"vendor", "bootstrap", "tidy", "lock"},
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
