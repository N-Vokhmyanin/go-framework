package zap_log

import (
	"github.com/N-Vokhmyanin/go-framework/logger"
	"go.uber.org/zap"
)

type zapLoggerAdapter struct {
	*zap.SugaredLogger
}

var _ logger.Logger = (*zapLoggerAdapter)(nil)

func NewZapLoggerAdapter(base *zap.Logger) logger.Logger {
	return &zapLoggerAdapter{base.Sugar()}
}

func (l *zapLoggerAdapter) With(args ...interface{}) logger.Logger {
	return &zapLoggerAdapter{l.SugaredLogger.With(args...)}
}

func (l *zapLoggerAdapter) WithOptions(opts ...logger.Option) logger.Logger {
	zapBase := l.SugaredLogger.Desugar()
	var logOpts logger.Options
	for _, opt := range opts {
		opt(&logOpts)
	}
	var zapOpts []zap.Option
	if logOpts.DisableStackTrace {
		zapOpts = append(zapOpts, zap.AddStacktrace(zapLevelEnablerNoneInstance))
	}
	return NewZapLoggerAdapter(zapBase.WithOptions(zapOpts...))
}
