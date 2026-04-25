package clientcore

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type ClientCore struct {
	scmProvider scm.Provider
}

func (cc *ClientCore) Initialize() error {
	scmProvider, err := scm.InitializeDefaultProvider()
	if err != nil {
		return seederr.Wrap(err)
	}
	cc.scmProvider = scmProvider
	return nil
}

func (cc *ClientCore) NdCut(args []string) error {
	return NdCut(cc.scmProvider, args)
}

func (cc *ClientCore) NdDev(args []string) error {
	return NdDev(cc.scmProvider, args)
}

func (cc *ClientCore) NdSetup(args []string) error {
	return NdSetup(args)
}

func (cc *ClientCore) NdShell(args []string) error {
	return NdShell(args)
}

func (cc *ClientCore) NdSubmit(args []string) error {
	return NdSubmit(cc.scmProvider, args)
}

func (cc *ClientCore) NdSync(args []string) error {
	return NdSync(cc.scmProvider, args)
}
