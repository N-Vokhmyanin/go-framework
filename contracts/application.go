package contracts

import (
	"github.com/urfave/cli/v2"
)

//goland:noinspection GoUnusedConst
const (
	EnvProduction  = "production"
	EnvDevelopment = "development"
)

type Application interface {
	CanBoot
	CanInit
	CanStart
	CanStop
	Container
	Env() string
	Name() string
	Provide(items ...Provider)
	Command(items ...*cli.Command)
	Run(args []string)
}

type Provider interface {
	Config(c ConfigSet)
	Boot(a Application)
	Register(a Application)
}
