package logger

type Options struct {
	DisableStackTrace bool
}

type Option func(o *Options)

// WithDisableStackTrace deprecated - always disabled
//
//goland:noinspection GoUnusedExportedFunction
func WithDisableStackTrace() Option {
	return func(o *Options) {
		o.DisableStackTrace = true
	}
}
