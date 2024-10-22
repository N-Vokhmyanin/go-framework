package database

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/application/ctxapp"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"github.com/N-Vokhmyanin/go-framework/database/migorm"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"github.com/N-Vokhmyanin/go-framework/logger/grpc/ctxlog"
	"github.com/N-Vokhmyanin/go-framework/utils/di"
	"gorm.io/gorm"
)

var _ migorm.Migrater = (*migraterService)(nil)
var _ contracts.CanStart = (*migraterService)(nil)

//goland:noinspection SpellCheckingInspection
type migraterService struct {
	cfg       migorm.Config
	app       contracts.Application
	logger    logger.Logger
	upOnStart bool
}

var _ migorm.Migrater = (*migraterService)(nil)

//goland:noinspection SpellCheckingInspection
func NewMigraterService(
	app contracts.Application,
	cfg migorm.Config,
	upOnStart bool,
) migorm.Migrater {
	lg := di.Get[logger.Logger](app).With(logger.WithComponent, "database.migrater")
	if cfg.Log == nil {
		cfg.Log = lg
	}
	return &migraterService{
		cfg:       cfg,
		app:       app,
		logger:    lg,
		upOnStart: upOnStart,
	}
}

//goland:noinspection SpellCheckingInspection
func (m *migraterService) migrater(connect bool) migorm.Migrater {
	var db *gorm.DB
	if connect {
		db = di.Get[Connection](m.app).DB()
	}

	return migorm.NewMigrater(db, &m.cfg)
}

func (m *migraterService) newContext(ctx context.Context) context.Context {
	ctx = ctxapp.Inject(ctx, m.app)
	ctx = ctxlog.ToContext(ctx, m.logger)

	return ctx
}

func (m *migraterService) Logger(ctx context.Context) migorm.Logger {
	return ctxlog.ExtractWithFallback(ctx, m.logger)
}

func (m *migraterService) UpMigrations(ctx context.Context) error {
	ctx = m.newContext(ctx)

	return m.migrater(true).UpMigrations(ctx)
}

func (m *migraterService) UpConcreteMigration(ctx context.Context, name string) error {
	ctx = m.newContext(ctx)

	return m.migrater(true).UpConcreteMigration(ctx, name)
}

func (m *migraterService) DownConcreteMigration(ctx context.Context, name string) error {
	ctx = m.newContext(ctx)

	return m.migrater(true).DownConcreteMigration(ctx, name)
}

func (m *migraterService) MakeFileMigration(ctx context.Context, name string) error {
	ctx = m.newContext(ctx)

	return m.migrater(false).MakeFileMigration(ctx, name)
}

func (m *migraterService) StartService() {
	if m.upOnStart {
		ctx := m.newContext(context.Background())
		err := m.UpMigrations(ctx)
		if err != nil {
			panic(err)
		}
	}
}
