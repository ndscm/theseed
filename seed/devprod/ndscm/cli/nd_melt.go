package main

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndMeltFlags struct {
	remove *seedflag.BoolFlag
	track  *seedflag.StringFlag
}

func parseNdMeltFlags(args []string) (ndMeltFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-melt")
	cmdFlags := ndMeltFlags{}
	cmdFlags.remove = cf.DefineBool("remove", false, "Remove the melt worktree")
	cmdFlags.track = cf.DefineString("track", "", "Local tracking branch for fork point search (default: origin/main)")
	cmdArgs, err := cf.Parse(args,
		seedflag.WithAnywhereFlag(true),
	)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndMelt(args []string) error {
	cmdFlags, cmdArgs, err := parseNdMeltFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	upstream := ""
	if len(cmdArgs) == 1 {
		upstream = cmdArgs[0]
	} else {
		return seederr.WrapErrorf("nd-melt usage: nd melt <upstream>")
	}
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdMelt(clientcore.NdMeltOptions{
		Remove:   cmdFlags.remove.Get(),
		Track:    cmdFlags.track.Get(),
		Upstream: upstream,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
