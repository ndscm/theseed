package main

import (
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndCutFlags struct {
	force *seedflag.BoolFlag
}

func parseNdCutFlags(args []string) (ndCutFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-cut")
	cmdFlags := ndCutFlags{}
	cmdFlags.force = cf.DefineBool("force", false, "Force cut even if the change branch already exists")
	cmdArgs, err := cf.Parse(args)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndCut(args []string) error {
	cmdFlags, cmdArgs, err := parseNdCutFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 2 {
		return seederr.WrapErrorf("nd-cut usage: nd cut [...flags] <feature-name> <ref>|<hash>")
	}
	featureName := strings.TrimSpace(cmdArgs[0])
	cutPoint := strings.TrimSpace(cmdArgs[1])
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdCut(clientcore.NdCutOptions{
		Force: cmdFlags.force.Get(),

		FeatureName: featureName,
		CutPoint:    cutPoint,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
