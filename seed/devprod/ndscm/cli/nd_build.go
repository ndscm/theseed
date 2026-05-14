package main

import (
	"runtime"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndBuildFlags struct {
	all *seedflag.BoolFlag
}

func parseNdBuildFlags(args []string) (ndBuildFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-build")
	cmdFlags := ndBuildFlags{}
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

func ndBuild(args []string) error {
	cmdFlags, cmdArgs, err := parseNdBuildFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 0 {
		return seederr.WrapErrorf("nd-build usage: ndscm build")
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

		Phases: []string{"vendor", "bootstrap", "build"},
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
