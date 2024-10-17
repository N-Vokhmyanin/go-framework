package database

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	health "github.com/N-Vokhmyanin/go-framework/health/contracts"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"go.uber.org/zap"
	grpcHealthV1 "google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"
	"sync"
	"time"
)

type gormConnection struct {
	sync.Mutex
	dialector gorm.Dialector
	db        *gorm.DB
	gormCfg   *gorm.Config
	log       logger.Logger
	connected bool
	closed    bool

	initFlag bool
	initLock sync.Mutex

	callbacks []Callback
}

var _ Connection = (*gormConnection)(nil)
var _ health.Service = (*gormConnection)(nil)
var _ contracts.CanStart = (*gormConnection)(nil)
var _ contracts.CanStop = (*gormConnection)(nil)

func NewGormConnection(dialector gorm.Dialector, gormCfg *gorm.Config, log logger.Logger) Connection {
	return &gormConnection{
		dialector: dialector,
		gormCfg:   gormCfg,
		log:       log,
	}
}

func (c *gormConnection) DB() *gorm.DB {
	c.Connect()
	return c.db
}

func (c *gormConnection) Connect() {
	c.initLock.Lock()
	defer c.initLock.Unlock()
	if c.initFlag {
		return
	}
	c.initFlag = true

	connect := func() {
		db, err := gorm.Open(c.dialector, c.gormCfg)
		if err != nil {
			c.log.Errorw("connect failed", zap.Error(err))
		} else {
			c.runCallbacks(db)
			c.db = db
			c.connected = true
			c.log.Infow("connected successful")
		}
	}

	ping := func() {
		sqlDB, err := c.db.DB()
		if err != nil {
			c.log.Errorw("get sql db failed", zap.Error(err))
			c.connected = false
			connect()
		}
		if err = sqlDB.Ping(); err != nil {
			c.log.Errorw("ping failed", zap.Error(err))
			c.connected = false
			connect()
		}
	}

	check := func() {
		c.Lock()
		defer c.Unlock()
		defer func() { _ = recover() }()
		if !c.closed {
			if !c.connected {
				connect()
			} else {
				ping()
			}
		}
	}

	go func() {
		for {
			check()
			time.Sleep(5 * time.Second)
		}
	}()

	// waiting for gormConnection
	for {
		if c.connected {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

}

func (c *gormConnection) Close() {
	c.Lock()
	defer c.Unlock()
	if c.db != nil && c.connected {
		sqlDB, err := c.db.DB()
		if err == nil {
			_ = sqlDB.Close()
			c.closed = true
		}
	}
}

func (c *gormConnection) IsConnected() bool {
	return c.connected
}

func (c *gormConnection) Register(cb Callback) {
	c.Lock()
	defer c.Unlock()
	c.callbacks = append(c.callbacks, cb)
	if c.connected && c.db != nil {
		cb(c.db)
	}
}

func (c *gormConnection) runCallbacks(db *gorm.DB) {
	for _, cb := range c.callbacks {
		cb(db)
	}
	c.log.Infow("callbacks registered")
}

func (c *gormConnection) HealthStatus(context.Context) grpcHealthV1.HealthCheckResponse_ServingStatus {
	return health.HealthStatusFromBool(c.connected)
}

func (c *gormConnection) StartService() {
	c.Connect()
}

func (c *gormConnection) StopService() {
	c.Close()
}
