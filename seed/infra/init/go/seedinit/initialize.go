package seedinit

import (
	"log/slog"

	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagVerbose = seedflag.DefineBool("verbose", false, "Enable verbose logging")

var packageInitializers = []func() error{}

func RegisterPackageInit(initializer func() error) {
	packageInitializers = append(packageInitializers, initializer)
}

func Initialize() error {
	err := seedflag.Parse()
	if err != nil {
		return err
	}
	if flagVerbose.Get() {
		seedlog.SetLevel(slog.LevelDebug)
	}
	for _, initializer := range packageInitializers {
		err := initializer()
		if err != nil {
			return err
		}
	}
	return nil
}
