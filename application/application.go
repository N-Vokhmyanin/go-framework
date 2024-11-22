package application

import (
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/health"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"github.com/N-Vokhmyanin/go-framework/logger/zap"
	"github.com/N-Vokhmyanin/go-framework/tracer"
	"github.com/N-Vokhmyanin/go-framework/utils/rfl"
	"github.com/urfave/cli/v2"
	"os"
	"sort"
	"sync"
)

type appInstance struct {
	name string
	env  string

	config    contracts.Config
	container contracts.Container
	providers []contracts.Provider
	commands  []*cli.Command

	loggerProvider logger.Provider
}

var _ contracts.Application = (*appInstance)(nil)

//goland:noinspection GoUnusedExportedFunction
func New(options ...Option) contracts.Application {
	a := &appInstance{
		config:         newConfig(),
		container:      newContainer(),
		loggerProvider: zap_log.NewZapProvider(),
	}
	for _, option := range options {
		option(a)
	}
	a.Provide(
		newAppProvider(a),
		a.loggerProvider,
		tracer.NewProvider(),
		health.NewProvider(),
	)
	return a
}

func (a *appInstance) Env() string {
	return a.env
}

func (a *appInstance) Name() string {
	return a.name
}

func (a *appInstance) Instances() []interface{} {
	return a.container.Instances()
}

func (a *appInstance) Singleton(resolver interface{}) {
	a.container.Singleton(resolver)
}

func (a *appInstance) Transient(resolver interface{}) {
	a.container.Transient(resolver)
}

func (a *appInstance) Make(receiver interface{}) []interface{} {
	return a.container.Make(receiver)
}

func (a *appInstance) Command(items ...*cli.Command) {
	a.commands = append(a.commands, items...)
}

func (a *appInstance) Provide(items ...contracts.Provider) {
	a.providers = append(a.providers, items...)
}

func (a *appInstance) BootService() {
	// Boot services
	for _, instance := range a.Instances() {
		if runnable, ok := instance.(contracts.CanBoot); ok && instance != a {
			runnable.BootService()
		}
	}
}

func (a *appInstance) InitService() {
	// Init services
	for _, instance := range a.Instances() {
		if runnable, ok := instance.(contracts.CanInit); ok && instance != a {
			runnable.InitService()
		}
	}
}

func (a *appInstance) StartService() {
	// Start services
	for _, instance := range a.Instances() {
		if runnable, ok := instance.(contracts.CanStart); ok && instance != a {
			runnable.StartService()
		}
	}
}

func (a *appInstance) StopService() {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	// Shutdown services
	for _, instance := range a.Instances() {
		if runnable, ok := instance.(contracts.CanStop); ok && instance != a {
			wg.Add(1)
			go func(r contracts.CanStop) {
				r.StopService()
				wg.Done()
			}(runnable)
		}
	}
}

func (a *appInstance) Run(args []string) {
	// Configure all providers
	for _, provider := range a.providers {
		providerKey := rfl.FullTypeName(provider)
		provider.Config(a.config.New(providerKey))
	}
	a.config.Parse(os.Args[1:])

	// Boot providers
	for _, provider := range a.providers {
		provider.Boot(a)
	}
	// Register providers
	for _, provider := range a.providers {
		provider.Register(a)
	}

	a.BootService()
	defer a.StopService()

	// Register commands
	cliApp := cli.NewApp()
	cliApp.Name = a.name
	cliApp.Commands = a.commands
	cliApp.Usage = a.name + " application"
	sort.Sort(cli.CommandsByName(cliApp.Commands))

	a.Make(
		func(log logger.Logger) {
			if err := cliApp.Run(args); err != nil {
				log.Error(err)
			}
		},
	)
}
