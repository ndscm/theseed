package seedflag

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type FlagDefinition interface {
	Usage() string
	flag.Value
}

type FlagItem struct {
	usage string
}

func (f *FlagItem) Usage() string {
	return f.usage
}

type BoolFlag struct {
	FlagItem
	value bool
}

func (f *BoolFlag) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return seederr.Wrap(err)
	}
	f.value = v
	return nil
}

func (f *BoolFlag) Get() bool {
	return f.value
}

func (f *BoolFlag) String() string {
	return strconv.FormatBool(f.value)
}

func (f *BoolFlag) IsBoolFlag() bool {
	return true
}

type StringFlag struct {
	FlagItem
	value string
}

func (f *StringFlag) Set(s string) error {
	f.value = s
	return nil
}

func (f *StringFlag) Get() string {
	return f.value
}

func (f *StringFlag) String() string {
	return f.value
}

type parseOptions struct {
}

type parseOption interface {
	applyParseOption(*parseOptions)
}

type globalParseOptions struct {
	parseOptions []parseOption

	envPrefix         string
	fallbackEnvPrefix string
}

type globalParseOption interface {
	applyGlobalParseOption(*globalParseOptions)
}

type CommandFlags struct {
	command string
	flags   map[string]FlagDefinition
}

func (cf *CommandFlags) DefineBool(name string, defaultValue bool, usage string) *BoolFlag {
	item := &BoolFlag{
		FlagItem: FlagItem{
			usage: usage,
		},
		value: defaultValue,
	}
	cf.flags[name] = item
	return item
}

func (cf *CommandFlags) DefineString(name string, defaultValue string, usage string) *StringFlag {
	item := &StringFlag{
		FlagItem: FlagItem{
			usage: usage,
		},
		value: defaultValue,
	}
	cf.flags[name] = item
	return item
}

func (cf *CommandFlags) Parse(args []string, opts ...parseOption) ([]string, error) {
	o := &parseOptions{}
	for _, opt := range opts {
		opt.applyParseOption(o)
	}
	s := flag.NewFlagSet(cf.command, flag.ContinueOnError)
	for name, item := range cf.flags {
		s.Var(item, name, item.Usage())
	}
	finalArgs := []string{}
	err := s.Parse(args)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	finalArgs = s.Args()
	return finalArgs, nil
}

func NewCommandFlags(command string) *CommandFlags {
	return &CommandFlags{
		command: command,
		flags:   map[string]FlagDefinition{},
	}
}

var globalFlags = NewCommandFlags(os.Args[0])

func DefineBool(name string, defaultValue bool, usage string) *BoolFlag {
	return globalFlags.DefineBool(name, defaultValue, usage)
}

func DefineString(name string, defaultValue string, usage string) *StringFlag {
	return globalFlags.DefineString(name, defaultValue, usage)
}

type withEnvPrefix struct {
	prefix string
}

func (w *withEnvPrefix) applyGlobalParseOption(o *globalParseOptions) {
	o.envPrefix = w.prefix
}

func WithEnvPrefix(prefix string) globalParseOption {
	return &withEnvPrefix{prefix: prefix}
}

type withFallbackEnvPrefix struct {
	prefix string
}

func (w *withFallbackEnvPrefix) applyGlobalParseOption(o *globalParseOptions) {
	o.fallbackEnvPrefix = w.prefix
}

func WithFallbackEnvPrefix(prefix string) globalParseOption {
	return &withFallbackEnvPrefix{prefix: prefix}
}

func Parse(opts ...globalParseOption) ([]string, error) {
	o := &globalParseOptions{}
	for _, opt := range opts {
		opt.applyGlobalParseOption(o)
	}

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if o.envPrefix != "" && strings.HasPrefix(parts[0], o.envPrefix) {
			pure := strings.TrimPrefix(parts[0], o.envPrefix)
			name := strings.ToLower(strings.TrimSpace(pure))
			value := strings.TrimSpace(parts[1])
			item, ok := globalFlags.flags[name]
			if !ok {
				seedlog.Warnf("Unknown env flag: %s=%s", parts[0], value)
				continue
			}
			err := item.Set(value)
			if err != nil {
				return nil, seederr.Wrap(err)
			}
		} else if o.fallbackEnvPrefix != "" && strings.HasPrefix(parts[0], o.fallbackEnvPrefix) {
			pure := strings.TrimPrefix(parts[0], o.fallbackEnvPrefix)
			name := strings.ToLower(strings.TrimSpace(pure))
			value := strings.TrimSpace(parts[1])
			item, ok := globalFlags.flags[name]
			if !ok {
				continue
			}
			seedlog.Infof("Loaded fallback env flag: %s=%s", parts[0], value)
			err := item.Set(value)
			if err != nil {
				return nil, seederr.Wrap(err)
			}
		}
	}

	args, err := globalFlags.Parse(os.Args[1:])
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	seedlog.Infof("Global flags: %+v", globalFlags.flags)
	return args, nil
}
