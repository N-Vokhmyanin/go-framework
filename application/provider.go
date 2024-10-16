package application

import (
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"os"
	"strings"
)

type appProvider struct {
	app *appInstance
}

var _ contracts.Provider = (*appProvider)(nil)

func newAppProvider(app *appInstance) contracts.Provider {
	return &appProvider{app}
}

func (p *appProvider) Config(c contracts.ConfigSet) {
	defaultAppName := "awesome"
	if len(os.Args) > 0 {
		pathParts := strings.Split(os.Args[0], "/")
		defaultAppName = pathParts[len(pathParts)-1]
	}
	c.StringVar(&p.app.name, "APP_NAME", defaultAppName, "application name")
	c.StringVar(&p.app.env, "APP_ENV", "development", "app environment")
}

func (p *appProvider) Boot(a contracts.Application) {
	a.Singleton(func() contracts.Application {
		return a
	})
	a.Singleton(func() contracts.Config {
		return p.app.config
	})
	a.Singleton(func() contracts.Dispatcher {
		return NewDispatcher()
	})
}

func (p *appProvider) Register(a contracts.Application) {
	a.Make(func(cfg contracts.Config, log logger.Logger) {
		a.Command(NewAppCommands(a, cfg, log)...)
	})
}
