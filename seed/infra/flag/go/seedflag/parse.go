package seedflag

import (
	"flag"
	"os"
	"strconv"
	"strings"
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
		return err
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

func Parse() error {
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			name := strings.ToLower(strings.TrimSpace(parts[0]))
			value := strings.TrimSpace(parts[1])
			if item, ok := configs[name]; ok {
				err := item.Set(value)
				if err != nil {
					return err
				}
			}
		}
	}
	for name, item := range configs {
		flag.Var(item, name, item.Usage())
	}
	flag.Parse()
	return nil
}
