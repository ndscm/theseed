package seedflag

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type ConfigDefinition interface {
	Usage() string
	flag.Value
}

type ConfigItem struct {
	usage string
}

func (c *ConfigItem) Usage() string {
	return c.usage
}

var configs = map[string]ConfigDefinition{}

type BoolConfig struct {
	ConfigItem
	value bool
}

func (d *BoolConfig) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return seederr.Wrap(err)
	}
	d.value = v
	return nil
}

func (d *BoolConfig) Get() bool {
	return d.value
}

func (d *BoolConfig) String() string {
	return strconv.FormatBool(d.value)
}

func (d *BoolConfig) IsBoolFlag() bool {
	return true
}

func DefineBool(name string, defaultValue bool, usage string) *BoolConfig {
	item := &BoolConfig{
		ConfigItem: ConfigItem{
			usage: usage,
		},
		value: defaultValue,
	}
	configs[name] = item
	return item
}

type StringConfig struct {
	ConfigItem
	value string
}

func (d *StringConfig) Set(s string) error {
	d.value = s
	return nil
}

func (d *StringConfig) Get() string {
	return d.value
}

func (d *StringConfig) String() string {
	return d.value
}

func DefineString(name string, defaultValue string, usage string) *StringConfig {
	item := &StringConfig{
		ConfigItem: ConfigItem{
			usage: usage,
		},
		value: defaultValue,
	}
	configs[name] = item
	return item
}

type parseOptions struct {
	envPrefix         string
	fallbackEnvPrefix string
}

type parseOption func(*parseOptions)

func WithEnvPrefix(prefix string) parseOption {
	return func(o *parseOptions) {
		o.envPrefix = prefix
	}
}

func WithFallbackEnvPrefix(prefix string) parseOption {
	return func(o *parseOptions) {
		o.fallbackEnvPrefix = prefix
	}
}

func Parse(opts ...parseOption) error {
	o := &parseOptions{}
	for _, opt := range opts {
		opt(o)
	}

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.HasPrefix(parts[0], o.envPrefix) {
			pure := strings.TrimPrefix(parts[0], o.envPrefix)
			name := strings.ToLower(strings.TrimSpace(pure))
			value := strings.TrimSpace(parts[1])
			item, ok := configs[name]
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
			item, ok := configs[name]
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

	for name, item := range configs {
		flag.Var(item, name, item.Usage())
	}
	flag.Parse()
	seedlog.Infof("flags: %+v", configs)
	return nil
}
