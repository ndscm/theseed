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

func (o *InitializeOptions) apply(opts *InitializeOptions) {
	*opts = *o
}

type initializeOption interface {
	apply(*InitializeOptions)
}

type anywhereFlagOption struct {
	value bool
}

func (o *anywhereFlagOption) apply(opts *InitializeOptions) {
	opts.AnywhereFlag = o.value
}

func WithAnywhereFlag(value bool) initializeOption {
	seedlog.Warnf("Deprecated, use seedinit.Initialize(&seedinit.InitializeOptions{...})")
	return &anywhereFlagOption{value: value}
}

type unknownFlagOption struct {
	value bool
}

func (o *unknownFlagOption) apply(opts *InitializeOptions) {
	opts.UnknownFlag = o.value
}

func WithUnknownFlag(value bool) initializeOption {
	seedlog.Warnf("Deprecated, use seedinit.Initialize(&seedinit.InitializeOptions{...})")
	return &unknownFlagOption{value: value}
}

type systemEnvOption struct {
	path string
}

func (o *systemEnvOption) apply(opts *InitializeOptions) {
	opts.SystemEnv = o.path
}

func WithSystemEnv(pathInConfigHome string) initializeOption {
	seedlog.Warnf("Deprecated, use seedinit.Initialize(&seedinit.InitializeOptions{...})")
	return &systemEnvOption{path: pathInConfigHome}
}

type userEnvOption struct {
	path string
}

func (o *userEnvOption) apply(opts *InitializeOptions) {
	opts.UserEnv = o.path
}

func WithUserEnv(pathInConfigHome string) initializeOption {
	seedlog.Warnf("Deprecated, use seedinit.Initialize(&seedinit.InitializeOptions{...})")
	return &userEnvOption{path: pathInConfigHome}
}

type ancestorEnvOption struct {
	fileName string
}

func (o *ancestorEnvOption) apply(opts *InitializeOptions) {
	opts.AncestorEnv = o.fileName
}

func WithAncestorEnv(envFileName string) initializeOption {
	seedlog.Warnf("Deprecated, use seedinit.Initialize(&seedinit.InitializeOptions{...})")
	return &ancestorEnvOption{fileName: envFileName}
}

type envPrefixOption struct {
	prefix string
}

func (o *envPrefixOption) apply(opts *InitializeOptions) {
	opts.EnvPrefix = o.prefix
}

func WithEnvPrefix(prefix string) initializeOption {
	seedlog.Warnf("Deprecated, use seedinit.Initialize(&seedinit.InitializeOptions{...})")
	return &envPrefixOption{prefix: prefix}
}

type fallbackEnvPrefixOption struct {
	prefix string
}

func (o *fallbackEnvPrefixOption) apply(opts *InitializeOptions) {
	opts.FallbackEnvPrefixes = append(opts.FallbackEnvPrefixes, o.prefix)
}

func WithFallbackEnvPrefix(prefix string) initializeOption {
	seedlog.Warnf("Deprecated, use seedinit.Initialize(&seedinit.InitializeOptions{...})")
	return &fallbackEnvPrefixOption{prefix: prefix}
}

// Use Initialize(&InitializeOptions{...}), avoid using deprecated With* options.
func Initialize(options ...initializeOption) ([]string, error) {
	opts := &InitializeOptions{}
	for _, opt := range options {
		if opt != nil {
			opt.apply(opts)
		}
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
