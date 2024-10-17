package database

import (
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/database/migorm"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"time"
)

type gormProvider struct {
	configs Configs
}

var _ contracts.Provider = (*gormProvider)(nil)

//goland:noinspection GoUnusedExportedFunction
func NewGormProvider(configs Configs) contracts.Provider {
	return &gormProvider{
		configs: configs,
	}
}

func (p *gormProvider) Config(c contracts.ConfigSet) {
	for _, cfg := range p.configs {
		c.StringVar(&cfg.DBHost, cfg.UpperPrefix()+"DB_HOST", "mysql", cfg.LowerPrefix()+"db host")
		c.StringVar(&cfg.DBPort, cfg.UpperPrefix()+"DB_PORT", "3306", cfg.LowerPrefix()+"db port")
		c.StringVar(&cfg.DBUser, cfg.UpperPrefix()+"DB_USER", "user", cfg.LowerPrefix()+"db username")
		c.StringVar(&cfg.DBPass, cfg.UpperPrefix()+"DB_PASS", "password", cfg.LowerPrefix()+"db password")
		c.StringVar(&cfg.DBName, cfg.UpperPrefix()+"DB_NAME", "database", cfg.LowerPrefix()+"db name")

		c.StringVar(&cfg.LogLevel, cfg.UpperPrefix()+"DB_LOG_LEVEL", "warn", "gorm log level")
		c.DurationVar(&cfg.SlowThreshold, cfg.UpperPrefix()+"DB_SLOW_THRESHOLD", 200*time.Millisecond, "slow threshold duration")
		c.BoolVar(&cfg.MigrationsRunOnStart, cfg.UpperPrefix()+"RUN_MIGRATIONS", false, "run migrations on start")
	}
}

func (p *gormProvider) Boot(a contracts.Application) {
	a.Singleton(NewConnectionRegistry)
	a.Singleton(func(r ConnectionRegistry) ConnectionPool { return r })

	defaultCfg := p.configs.Default()
	if defaultCfg != nil {
		a.Singleton(func(r ConnectionRegistry) Connection {
			return r.Default()
		})
		a.Singleton(func(conn Connection, log logger.Logger) migorm.Migrater {
			return NewMigraterService(
				a,
				migorm.Config{
					TableName:     defaultCfg.MigrationsTableName,
					MigrationsDir: defaultCfg.MigrationsDir,
				},
				defaultCfg.MigrationsRunOnStart,
			)
		})
	}
}

func (p *gormProvider) Register(a contracts.Application) {

	a.Make(func(
		connRegistry ConnectionRegistry,
		log logger.Logger,
	) {
		for _, cfg := range p.configs {
			dialector := cfg.Dialector()
			if dialector == nil {
				log.Error("unknown database driver: " + cfg.Driver)
				continue
			}
			log = log.With(logger.WithComponent, "database."+dialector.Name())
			gormCfg := &gorm.Config{}
			if gormCfg.Logger == nil {
				level, ok := GormLoggerLevels[cfg.LogLevel]
				if !ok {
					level = gormLogger.Warn
				}
				logCfg := gormLogger.Config{
					SlowThreshold: cfg.SlowThreshold,
					LogLevel:      level,
				}
				gormCfg.Logger = NewGormLoggerAdapter(log, logCfg)
			}
			conn := NewGormConnection(dialector, gormCfg, log)
			connRegistry.Register(cfg.Name, conn)
		}
	})

	a.Make(func(mig migorm.Migrater) {
		a.Command(NewMigrateCommands(a, mig)...)
	})
}
