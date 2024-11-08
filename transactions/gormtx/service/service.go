package service

import (
	"context"
	"sync"

	"github.com/N-Vokhmyanin/go-framework/database"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"github.com/N-Vokhmyanin/go-framework/logger/grpc/ctxlog"
	"github.com/N-Vokhmyanin/go-framework/transactions/gormtx"
	"gorm.io/gorm"
)

type service struct {
	sync.Mutex
	conn database.Connection
	log  logger.Logger
}

func NewService(conn database.Connection, log logger.Logger) gormtx.Service {
	return &service{
		conn: conn,
		log:  log,
	}
}

func (s *service) Start(ctx context.Context) (tx gormtx.Transaction, newCtx context.Context) {
	s.Lock()
	defer s.Unlock()

	defer func() {
		if btx, ok := tx.(interface{ Begin() }); ok {
			btx.Begin()
		}
	}()

	tx = gormtx.ExtractFromContext(ctx)
	if tx != nil {
		return tx, ctx
	}

	log := ctxlog.ExtractWithFallback(ctx, s.log)

	tx = &transaction{
		newDB: func() *gorm.DB {
			return s.conn.DB().Session(&gorm.Session{
				NewDB:       true,
				Initialized: true,
				Context:     ctx,
			})
		},
		log: log,
	}

	newCtx = gormtx.InjectToContext(ctx, tx)

	return tx, newCtx
}

func (s *service) DB(ctx context.Context) (db *gorm.DB) {
	s.Lock()
	defer s.Unlock()

	tx := gormtx.ExtractFromContext(ctx)
	if tx != nil {
		db = tx.DB()
	}

	if db != nil {
		return db
	}

	return s.conn.DB().Session(&gorm.Session{
		NewDB:       true,
		Initialized: true,
		Context:     ctx,
	})
}
