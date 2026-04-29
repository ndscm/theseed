package main

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndDevFlags struct {
	remove *seedflag.BoolFlag
	track  *seedflag.StringFlag
}

func parseNdDevFlags(args []string) (ndDevFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-dev")
	cmdFlags := ndDevFlags{}
	cmdFlags.remove = cf.DefineBool("remove", false, "Remove the dev worktree")
	cmdFlags.track = cf.DefineString("track", "", "Remote branch to track (e.g. origin/main); only valid when creating a new dev worktree, must not be provided if the dev worktree already exists")
	cmdArgs, err := cf.Parse(args,
		seedflag.WithAnywhereFlag(true),
	)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndDev(args []string) error {
	cmdFlags, cmdArgs, err := parseNdDevFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	focus := ""
	if len(cmdArgs) == 0 {
		// pass
	} else if len(cmdArgs) == 1 {
		focus = cmdArgs[0]
	} else {
		return seederr.WrapErrorf("nd-dev usage: nd dev [<focus-area>]")
	}
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdDev(clientcore.NdDevOptions{
		Remove: cmdFlags.remove.Get(),
		Track:  cmdFlags.track.Get(),
		Focus:  focus,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
