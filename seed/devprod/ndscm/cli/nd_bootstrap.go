package main

import (
	"runtime"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndBootstrapFlags struct {
	all *seedflag.BoolFlag
}

func parseNdBootstrapFlags(args []string) (ndBootstrapFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-bootstrap")
	cmdFlags := ndBootstrapFlags{}
	cmdFlags.all = cf.DefineBool("all", false, "Bootstrap all files in the repository")
	cmdArgs, err := cf.Parse(args,
		seedflag.WithAnywhereFlag(true),
	)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndBootstrap(args []string) error {
	cmdFlags, cmdArgs, err := parseNdBootstrapFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 0 {
		return seederr.WrapErrorf("nd-bootstrap usage: ndscm bootstrap")
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

		Phases: []string{"bootstrap"},
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
