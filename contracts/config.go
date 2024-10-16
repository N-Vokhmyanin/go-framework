package contracts

import "time"

type Config interface {
	New(provider string) ConfigSet
	Parse(arguments []string)
	Visit(fn func(f ConfigFlag))
}

type ConfigSet interface {
	BoolVar(p *bool, name string, value bool, usage string)
	IntVar(p *int, name string, value int, usage string)
	Int64Var(p *int64, name string, value int64, usage string)
	UintVar(p *uint, name string, value uint, usage string)
	Uint64Var(p *uint64, name string, value uint64, usage string)
	StringVar(p *string, name string, value string, usage string)
	Float64Var(p *float64, name string, value float64, usage string)
	DurationVar(p *time.Duration, name string, value time.Duration, usage string)
}

type ConfigFlag interface {
	Name() string
	Usage() string
	Value() string
	Default() string
	Provider() string
}
