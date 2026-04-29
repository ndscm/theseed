package main

import (
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndApplyFlags struct {
	remote *seedflag.StringFlag
	owner  *seedflag.StringFlag
}

func parseNdApplyFlags(args []string) (ndApplyFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-apply")
	cmdFlags := ndApplyFlags{}
	cmdFlags.remote = cf.DefineString("remote", "origin", "Remote identifier for fetching the branch")
	cmdFlags.owner = cf.DefineString("owner", "", "Owner of the remote change branch")
	cmdArgs, err := cf.Parse(args)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndApply(args []string) error {
	cmdFlags, cmdArgs, err := parseNdApplyFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 1 {
		return seederr.WrapErrorf("nd-apply usage: nd apply [...flags] <feature-name>")
	}
	featureName := strings.TrimSpace(cmdArgs[0])
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdApply(clientcore.NdApplyOptions{
		Remote: cmdFlags.remote.Get(),
		Owner:  cmdFlags.owner.Get(),

		FeatureName: featureName,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
