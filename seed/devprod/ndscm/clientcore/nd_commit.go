package clientcore

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type NdCommitOptions struct {
	Message string
}

func NdCommit(scmProvider scm.Provider, options NdCommitOptions) error {
	if options.Message == "" {
		return seederr.WrapErrorf("nd-commit requires a non-empty message")
	}
	err := scmProvider.CreateCommit("", options.Message)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
