package main

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

func parseNdContinueFlags(args []string) ([]string, error) {
	cf := seedflag.NewCommandFlags("nd-continue")
	cmdArgs, err := cf.Parse(args)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdArgs, nil
}

func ndContinue(args []string) error {
	cmdArgs, err := parseNdContinueFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 0 {
		return seederr.WrapErrorf("nd-continue usage: nd continue")
	}
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdContinue(clientcore.NdContinueOptions{})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
