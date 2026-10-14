package seedinit

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ndscm/theseed/seed/infra/buildinfo/go/buildinfo"
	"github.com/ndscm/theseed/seed/infra/dotenv/go/dotenv"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagVerbose = seedflag.DefineBool("verbose", false, "Enable verbose logging")
var flagVersion = seedflag.DefineBool("version", false, "Print version and exit")

var packageInitializers = []func() error{}

func RegisterPackageInit(initializer func() error) {
	packageInitializers = append(packageInitializers, initializer)
}

type InitializeOptions struct {
	seedflag.GlobalParseOptions

	// SystemEnv is the env file path relative to /etc.
	SystemEnv string

	// UserEnv is the env file path relative to the user's config directory.
	UserEnv string

	// AncestorEnv is the env filename to search for in the current directory and its ancestors.
	AncestorEnv string
}

// Use Initialize(&InitializeOptions{...}), avoid using deprecated With* options.
func Initialize(opts *InitializeOptions) ([]string, error) {
	if opts == nil {
		opts = &InitializeOptions{}
	}

	if opts.SystemEnv != "" {
		err := loadSystemEnv(opts.SystemEnv)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
	}
	if opts.UserEnv != "" {
		err := loadUserEnv(opts.UserEnv)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
	}
	if opts.AncestorEnv != "" {
		err := dotenv.LoadAncestor(".", opts.AncestorEnv)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
	}

	args, err := seedflag.Parse(&opts.GlobalParseOptions)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if flagVersion.Get() {
		info := buildinfo.GetBuildInfo()
		fmt.Println(info.BuildBrief)
		// Exit here to mirror the stdlib flag.CommandLine behavior,
		os.Exit(0)
		return nil, seederr.WrapErrorf("print version")
	}
	if flagVerbose.Get() {
		seedlog.SetLevel(slog.LevelDebug)
	}
	for _, initializer := range packageInitializers {
		err := initializer()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
	}
	return args, nil
}
