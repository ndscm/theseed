package main

import (
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndSubmitFlags struct {
	remote *seedflag.StringFlag
}

func parseNdSubmitFlags(args []string) (ndSubmitFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-submit")
	cmdFlags := ndSubmitFlags{}
	cmdFlags.remote = cf.DefineString("remote", "origin", "Remote identifier for submitting the branch")
	cmdArgs, err := cf.Parse(args)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndSubmit(args []string) error {
	cmdFlags, cmdArgs, err := parseNdSubmitFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 1 {
		return seederr.WrapErrorf("nd-submit usage: nd submit [...flags] <feature-name>")
	}
	featureName := strings.TrimSpace(cmdArgs[0])
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdSubmit(clientcore.NdSubmitOptions{
		Remote: cmdFlags.remote.Get(),

		FeatureName: featureName,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
