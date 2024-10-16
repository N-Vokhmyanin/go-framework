package application

import (
	"fmt"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/namsral/flag"
)

type defaultConfig struct {
	configs map[string]*flag.FlagSet
}

var _ contracts.Config = (*defaultConfig)(nil)

func newConfig() contracts.Config {
	return &defaultConfig{
		configs: make(map[string]*flag.FlagSet),
	}
}

func (c *defaultConfig) New(provider string) contracts.ConfigSet {
	if _, ok := c.configs[provider]; ok {
		panic(fmt.Errorf(`flag set with name "%s" already registered`, provider))
	}
	c.configs[provider] = flag.NewFlagSet(provider, flag.PanicOnError)
	c.configs[provider].String(flag.DefaultConfigFlagname, "", "path to config file")
	return c.configs[provider]
}

func (c *defaultConfig) Parse(arguments []string) {
	for group, set := range c.configs {
		if err := set.Parse(arguments); err != nil {
			panic(fmt.Errorf(`parse "%s" error: %s`, group, err))
		}
	}
}

func (c *defaultConfig) Visit(fn func(f contracts.ConfigFlag)) {
	for group, set := range c.configs {
		set.VisitAll(func(f *flag.Flag) {
			fn(&defaultConfigFlag{
				name:     f.Name,
				usage:    f.Usage,
				value:    f.Value.String(),
				defValue: f.DefValue,
				provider: group,
			})
		})
	}
}

type defaultConfigFlag struct {
	name     string
	usage    string
	value    string
	defValue string
	provider string
}

var _ contracts.ConfigSet = (*flag.FlagSet)(nil)
var _ contracts.ConfigFlag = (*defaultConfigFlag)(nil)

func (f *defaultConfigFlag) Name() string {
	return f.name
}

func (f *defaultConfigFlag) Usage() string {
	return f.usage
}

func (f *defaultConfigFlag) Value() string {
	return f.value
}

func (f *defaultConfigFlag) Default() string {
	return f.defValue
}

func (f *defaultConfigFlag) Provider() string {
	return f.provider
}
