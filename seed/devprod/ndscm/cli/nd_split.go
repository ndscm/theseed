package main

import (
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

func parseNdSplitFlags(args []string) ([]string, error) {
	cf := seedflag.NewCommandFlags("nd-split")
	cmdArgs, err := cf.Parse(args)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdArgs, nil
}

func ndSplit(args []string) error {
	cmdArgs, err := parseNdSplitFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 1 {
		return seederr.WrapErrorf("nd-split usage: nd split <ref>|<hash>")
	}
	commit := strings.TrimSpace(cmdArgs[0])
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdSplit(clientcore.NdSplitOptions{
		Commit: commit,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
