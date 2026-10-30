package main

import (
	"strconv"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndSubmitFlags struct {
	force  *seedflag.BoolFlag
	remote *seedflag.StringFlag

	owner *seedflag.StringFlag

	maxCommits *seedflag.StringFlag
}

func parseNdSubmitFlags(args []string) (ndSubmitFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-submit")
	cmdFlags := ndSubmitFlags{}
	cmdFlags.force = cf.DefineBool("force", false, "Force cut even if the change branch already exists")
	cmdFlags.remote = cf.DefineString("remote", "origin", "Remote identifier for submitting the branch")
	cmdFlags.owner = cf.DefineString("owner", "", "Owner identifier for submitting the branch")
	cmdFlags.maxCommits = cf.DefineString("max-commits", "0", "Cap the number of commits applied to the submit branch")
	cmdArgs, err := cf.Parse(args, &seedflag.ParseOptions{
		AnywhereFlag: true,
	})
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
	if len(cmdArgs) < 1 || len(cmdArgs) > 2 {
		return seederr.WrapErrorf("nd-submit usage: nd submit [...flags] <feature-name> [<ref>|<hash>]")
	}
	featureName := strings.TrimSpace(cmdArgs[0])
	cutPoint := ""
	if len(cmdArgs) == 2 {
		cutPoint = strings.TrimSpace(cmdArgs[1])
	}
	maxCommits, err := strconv.Atoi(strings.TrimSpace(cmdFlags.maxCommits.Get()))
	if err != nil || maxCommits < 0 {
		return seederr.WrapErrorf("nd-submit: --max-commits must be a non-negative integer")
	}
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdSubmit(clientcore.NdSubmitOptions{
		Force:  cmdFlags.force.Get(),
		Remote: cmdFlags.remote.Get(),
		Owner:  cmdFlags.owner.Get(),

		MaxCommits: maxCommits,

		FeatureName: featureName,
		CutPoint:    cutPoint,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
