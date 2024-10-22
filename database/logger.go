package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/N-Vokhmyanin/go-framework/logger"
	"github.com/N-Vokhmyanin/go-framework/logger/grpc/ctxlog"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

const (
	traceStr     = "%s\n[%.3fms] [rows:%v] %s"
	traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
	traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
)

var GormLoggerLevels = map[string]gormLogger.LogLevel{
	"info":   gormLogger.Info,
	"warn":   gormLogger.Warn,
	"error":  gormLogger.Error,
	"silent": gormLogger.Silent,
}

type gormLoggerAdapter struct {
	gormLogger.Config
	logger logger.Logger
}

var _ gormLogger.Interface = (*gormLoggerAdapter)(nil)

func NewGormLoggerAdapter(lg logger.Logger, cfg gormLogger.Config) gormLogger.Interface {
	return &gormLoggerAdapter{Config: cfg, logger: lg}
}

func (adapter *gormLoggerAdapter) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newAdapter := *adapter
	newAdapter.LogLevel = level

	return &newAdapter
}

func (adapter *gormLoggerAdapter) getLogger(ctx context.Context) logger.Logger {
	return ctxlog.ExtractWithFallback(ctx, adapter.logger)
}

func (adapter *gormLoggerAdapter) Info(ctx context.Context, msg string, args ...interface{}) {
	if adapter.LogLevel >= gormLogger.Info {
		adapter.getLogger(ctx).Infof(msg, args...)
	}
}

func (adapter *gormLoggerAdapter) Warn(ctx context.Context, msg string, args ...interface{}) {
	if adapter.LogLevel >= gormLogger.Warn {
		adapter.getLogger(ctx).Warnf(msg, args...)
	}
}

func (adapter *gormLoggerAdapter) Error(ctx context.Context, msg string, args ...interface{}) {
	if adapter.LogLevel >= gormLogger.Error {
		adapter.getLogger(ctx).Errorf(msg, args...)
	}
}

func (adapter *gormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	lg := adapter.getLogger(ctx)
	elapsed := time.Since(begin)

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && adapter.LogLevel >= gormLogger.Error:
		sql, rows := fc()
		if rows == -1 {
			lg.Errorf(traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			lg.Errorf(traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > adapter.SlowThreshold && adapter.SlowThreshold != 0 && adapter.LogLevel >= gormLogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", adapter.SlowThreshold)
		if rows == -1 {
			lg.Warnf(traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			lg.Warnf(traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case adapter.LogLevel >= gormLogger.Info:
		sql, rows := fc()
		if rows == -1 {
			lg.Infof(traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			lg.Infof(traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}
