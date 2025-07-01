package seedinit

import (
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var packageInitializers = []func() error{}

func RegisterPackageInit(initializer func() error) {
	packageInitializers = append(packageInitializers, initializer)
}

func Initialize() error {
	err := seedflag.Parse()
	if err != nil {
		return err
	}
	for _, initializer := range packageInitializers {
		err := initializer()
		if err != nil {
			return err
		}
	}
	return nil
}
