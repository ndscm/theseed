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

var globalFlags = map[string]FlagDefinition{}

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

func DefineBool(name string, defaultValue bool, usage string) *BoolFlag {
	item := &BoolFlag{
		FlagItem: FlagItem{
			usage: usage,
		},
		value: defaultValue,
	}
	globalFlags[name] = item
	return item
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

func DefineString(name string, defaultValue string, usage string) *StringFlag {
	item := &StringFlag{
		FlagItem: FlagItem{
			usage: usage,
		},
		value: defaultValue,
	}
	globalFlags[name] = item
	return item
}

type globalParseOptions struct {
	envPrefix         string
	fallbackEnvPrefix string
}

type globalParseOption func(*globalParseOptions)

func WithEnvPrefix(prefix string) globalParseOption {
	return func(o *globalParseOptions) {
		o.envPrefix = prefix
	}
}

func WithFallbackEnvPrefix(prefix string) globalParseOption {
	return func(o *globalParseOptions) {
		o.fallbackEnvPrefix = prefix
	}
}

func Parse(opts ...globalParseOption) error {
	o := &globalParseOptions{}
	for _, opt := range opts {
		opt(o)
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
			item, ok := globalFlags[name]
			if !ok {
				if o.envPrefix != "" {
					seedlog.Warnf("Unknown env flag: %s=%s", parts[0], value)
				}
				continue
			}
			err := item.Set(value)
			if err != nil {
				return seederr.Wrap(err)
			}
		} else if o.fallbackEnvPrefix != "" && strings.HasPrefix(parts[0], o.fallbackEnvPrefix) {
			pure := strings.TrimPrefix(parts[0], o.fallbackEnvPrefix)
			name := strings.ToLower(strings.TrimSpace(pure))
			value := strings.TrimSpace(parts[1])
			item, ok := globalFlags[name]
			if !ok {
				continue
			}
			seedlog.Infof("Loaded fallback env flag: %s=%s", parts[0], value)
			err := item.Set(value)
			if err != nil {
				return seederr.Wrap(err)
			}
		}
	}

	for name, item := range globalFlags {
		flag.Var(item, name, item.Usage())
	}
	flag.Parse()
	seedlog.Infof("Global flags: %+v", globalFlags)
	return nil
}
