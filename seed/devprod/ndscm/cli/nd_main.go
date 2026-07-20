package main

import (
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndMainFlags struct {
	remove *seedflag.BoolFlag
	orphan *seedflag.StringFlag
}

func parseNdMainFlags(args []string) (ndMainFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-main")
	cmdFlags := ndMainFlags{}
	cmdFlags.remove = cf.DefineBool("remove", false, "Remove the area main worktree")
	cmdFlags.orphan = cf.DefineString("orphan", "", "Create the area branch as an orphan branch with no history, rooted at an empty commit carrying this message")
	cmdArgs, err := cf.Parse(args,
		seedflag.WithAnywhereFlag(true),
	)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndMain(args []string) error {
	cmdFlags, cmdArgs, err := parseNdMainFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	area := ""
	start := ""
	if len(cmdArgs) == 0 {
		// pass
	} else if len(cmdArgs) == 1 {
		area = strings.TrimSpace(cmdArgs[0])
	} else if len(cmdArgs) == 2 {
		area = strings.TrimSpace(cmdArgs[0])
		start = strings.TrimSpace(cmdArgs[1])
	} else {
		return seederr.WrapErrorf("nd-main usage: nd main [...flags] [<area> [<start-point>]]")
	}
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdMain(clientcore.NdMainOptions{
		Remove: cmdFlags.remove.Get(),

		Orphan: cmdFlags.orphan.Get(),

		Area:  area,
		Start: start,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
