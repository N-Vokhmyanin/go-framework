package zap_log

import (
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"go.uber.org/zap"
)

var emptyLevel = zap.AtomicLevel{}

type config struct {
	Log *zap.Config `json:"log" yaml:"log"`
}

type provider struct {
	level   string
	cfgFile string

	config *zap.Config

	zap *zap.Logger
	log logger.Logger
}

var _ logger.Provider = (*provider)(nil)

func NewZapProvider() logger.Provider {
	return &provider{}
}

func (p *provider) WithConfig(cfg *zap.Config) *provider {
	p.config = cfg
	return p
}

func (p *provider) Config(c contracts.ConfigSet) {
	c.StringVar(&p.level, "LOG_LEVEL", "", "min log level")
	c.StringVar(&p.cfgFile, "LOG_CONFIG_FILE", "", "log config file")
}

func (p *provider) Boot(a contracts.Application) {
	a.Singleton(
		func(cfg *zap.Config) *zap.Logger {
			isDev := a.Env() != contracts.EnvProduction
			return p.getZapLogger(isDev, cfg)
		},
	)
	a.Singleton(
		func(zapLogger *zap.Logger) logger.Logger {
			p.log = NewZapLoggerAdapter(zapLogger)
			return p.log
		},
	)
}

func (p *provider) Register(contracts.Application) {
	// nothing
}
