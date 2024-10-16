package zap_log

import (
	"go.uber.org/zap/zapcore"
)

type zapLevelEnablerNone struct {
}

var _ zapcore.LevelEnabler = (*zapLevelEnablerNone)(nil)
var zapLevelEnablerNoneInstance = &zapLevelEnablerNone{}

func (zapLevelEnablerNone) Enabled(zapcore.Level) bool {
	return false
}
