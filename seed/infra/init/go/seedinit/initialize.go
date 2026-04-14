package seedinit

import (
	"log/slog"

	"github.com/ndscm/theseed/seed/infra/dotenv/go/dotenv"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagVerbose = seedflag.DefineBool("verbose", false, "Enable verbose logging")

var packageInitializers = []func() error{}

func RegisterPackageInit(initializer func() error) {
	packageInitializers = append(packageInitializers, initializer)
}

type initializeOptions struct {
	// env file locators
	systemEnvPath       string
	userEnvPath         string
	ancestorEnvFileName string

	envPrefix         string
	fallbackEnvPrefix string
}

type initializeOption func(*initializeOptions)

func WithSystemEnv(pathInConfigHome string) initializeOption {
	return func(o *initializeOptions) {
		o.systemEnvPath = pathInConfigHome
	}
}

func WithUserEnv(pathInConfigHome string) initializeOption {
	return func(o *initializeOptions) {
		o.userEnvPath = pathInConfigHome
	}
}

func WithAncestorEnv(envFileName string) initializeOption {
	return func(o *initializeOptions) {
		o.ancestorEnvFileName = envFileName
	}
}

func WithEnvPrefix(prefix string) initializeOption {
	return func(o *initializeOptions) {
		o.envPrefix = prefix
	}
}

func WithFallbackEnvPrefix(prefix string) initializeOption {
	return func(o *initializeOptions) {
		o.fallbackEnvPrefix = prefix
	}
}

func Initialize(opts ...initializeOption) error {
	o := &initializeOptions{}
	for _, opt := range opts {
		opt(o)
	}

	if o.systemEnvPath != "" {
		err := loadSystemEnv(o.systemEnvPath)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	if o.userEnvPath != "" {
		err := loadUserEnv(o.userEnvPath)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	if o.ancestorEnvFileName != "" {
		err := dotenv.LoadAncestor(".", o.ancestorEnvFileName)
		if err != nil {
			return seederr.Wrap(err)
		}
	}

	err := seedflag.Parse(
		seedflag.WithEnvPrefix(o.envPrefix),
		seedflag.WithFallbackEnvPrefix(o.fallbackEnvPrefix),
	)
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
