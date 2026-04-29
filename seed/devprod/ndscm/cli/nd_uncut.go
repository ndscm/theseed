package main

import (
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

func ndUncut(args []string) error {
	seedflag.Finalize(args)
	if len(args) != 1 {
		return seederr.WrapErrorf("nd-uncut usage: nd uncut <feature-name>")
	}
	featureName := strings.TrimSpace(args[0])
	cc := &clientcore.ClientCore{}
	err := cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdUncut(clientcore.NdUncutOptions{
		FeatureName: featureName,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
