package main

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndFormatFlags struct {
	all *seedflag.BoolFlag
}

func parseNdFormatFlags(args []string) (ndFormatFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-format")
	cmdFlags := ndFormatFlags{}
	cmdFlags.all = cf.DefineBool("all", false, "Format all files instead of only dirty files")
	cmdArgs, err := cf.Parse(args)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndFormat(args []string) error {
	cmdFlags, cmdArgs, err := parseNdFormatFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 0 {
		return seederr.WrapErrorf("nd-format usage: ndscm format")
	}
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdFormat(clientcore.NdFormatOptions{
		All: cmdFlags.all.Get(),
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
