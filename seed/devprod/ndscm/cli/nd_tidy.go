package main

import (
	"runtime"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndTidyFlags struct {
	all *seedflag.BoolFlag
}

func parseNdTidyFlags(args []string) (ndTidyFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-tidy")
	cmdFlags := ndTidyFlags{}
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

func ndTidy(args []string) error {
	cmdFlags, cmdArgs, err := parseNdTidyFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 0 {
		return seederr.WrapErrorf("nd-tidy usage: ndscm tidy")
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

		Phases: []string{"vendor", "bootstrap", "tidy"},
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
