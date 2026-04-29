package main

import (
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

func ndCut(args []string) error {
	seedflag.Finalize(args)
	if len(args) != 2 {
		return seederr.WrapErrorf("nd-cut usage: nd cut <feature-name> <ref>|<hash>")
	}
	featureName := strings.TrimSpace(args[0])
	cutPoint := strings.TrimSpace(args[1])
	cc := &clientcore.ClientCore{}
	err := cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdCut(clientcore.NdCutOptions{
		FeatureName: featureName,
		CutPoint:    cutPoint,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
