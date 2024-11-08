package provider

import (
	fw "github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/transactions"
	"github.com/N-Vokhmyanin/go-framework/transactions/gormtx"
	"github.com/N-Vokhmyanin/go-framework/transactions/gormtx/service"
)

type provider struct{}

func New() fw.Provider {
	return &provider{}
}

func (p *provider) Config(_ fw.ConfigSet) {
	// nothing
}

func (p *provider) Boot(a fw.Application) {
	a.Singleton(service.NewService)
	a.Singleton(func(svc gormtx.Service) transactions.Service {
		return &adapter{
			base: svc,
		}
	})
}

func (p *provider) Register(_ fw.Application) {
	// nothing
}
