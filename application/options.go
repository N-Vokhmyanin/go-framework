package application

import (
	"github.com/N-Vokhmyanin/go-framework/logger"
)

type Option func(a *appInstance)

//goland:noinspection GoUnusedExportedFunction
func LoggerOption(p logger.Provider) Option {
	return func(a *appInstance) {
		a.loggerProvider = p
	}
}
