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

func (cc *ClientCore) NdCut(options NdCutOptions) error {
	return NdCut(cc.scmProvider, options)
}

func (cc *ClientCore) NdDev(options NdDevOptions) error {
	return NdDev(cc.scmProvider, options)
}

func (cc *ClientCore) NdSetup(options NdSetupOptions) error {
	return NdSetup(options)
}

func (cc *ClientCore) NdShell(options NdShellOptions) error {
	return NdShell(options)
}

func (cc *ClientCore) NdSubmit(options NdSubmitOptions) error {
	return NdSubmit(cc.scmProvider, options)
}

func (cc *ClientCore) NdSync(options NdSyncOptions) error {
	return NdSync(cc.scmProvider, options)
}
