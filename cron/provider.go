package cron

import (
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/logger"
)

type cronProvider struct {
	enabled bool
}

var _ contracts.Provider = (*cronProvider)(nil)

//goland:noinspection GoUnusedExportedFunction
func NewCronProvider() contracts.Provider {
	return &cronProvider{}
}

func (p *cronProvider) Config(c contracts.ConfigSet) {
	c.BoolVar(&p.enabled, "CRON_ENABLED", false, "enable cron tasks")
}

func (p *cronProvider) Boot(a contracts.Application) {
	a.Singleton(func(log logger.Logger) Service {
		return NewGronService(p.enabled, log)
	})
}

func (p *cronProvider) Register(a contracts.Application) {
	a.Make(func(service Service, log logger.Logger) {
		a.Command(NewCronCommands(a, service, log)...)
	})
}
