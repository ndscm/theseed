package clientcore

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type NdAmendOptions struct {
	Break   string
	Migrate string
}

func NdAmend(scmProvider scm.Provider, options NdAmendOptions) error {
	if options.Break == "" && options.Migrate == "" {
		return seederr.WrapErrorf("nd-amend requires at least one of --break or --migrate")
	}
	if options.Break != "" {
		if options.Break == "lock" && options.Migrate == "" {
			options.Migrate = "update lock"
		}
		if options.Break == "melt" && options.Migrate == "" {
			options.Migrate = "drop"
		}
		err := scmProvider.AmendHeadCommit("Break", options.Break)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	if options.Migrate != "" {
		err := scmProvider.AmendHeadCommit("Migrate", options.Migrate)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	return nil
}
