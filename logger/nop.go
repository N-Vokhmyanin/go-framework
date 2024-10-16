package logger

type nopLogger struct{}

var _ Logger = (*nopLogger)(nil)
var nopLoggerInstance = &nopLogger{}

func GetNopLogger() Logger {
	return nopLoggerInstance
}

func (n *nopLogger) Debug(...interface{}) {}
func (n *nopLogger) Info(...interface{})  {}
func (n *nopLogger) Warn(...interface{})  {}
func (n *nopLogger) Error(...interface{}) {}
func (n *nopLogger) Panic(...interface{}) {}
func (n *nopLogger) Fatal(...interface{}) {}

//goland:noinspection SpellCheckingInspection
func (n *nopLogger) Debugf(string, ...interface{}) {}

//goland:noinspection SpellCheckingInspection
func (n *nopLogger) Infof(string, ...interface{}) {}

//goland:noinspection SpellCheckingInspection
func (n *nopLogger) Warnf(string, ...interface{}) {}

//goland:noinspection SpellCheckingInspection
func (n *nopLogger) Errorf(string, ...interface{}) {}

//goland:noinspection SpellCheckingInspection
func (n *nopLogger) Panicf(string, ...interface{}) {}
func (n *nopLogger) Fatalf(string, ...interface{}) {}

//goland:noinspection SpellCheckingInspection
func (n *nopLogger) Debugw(string, ...interface{}) {}
func (n *nopLogger) Infow(string, ...interface{})  {}

//goland:noinspection SpellCheckingInspection
func (n *nopLogger) Warnw(string, ...interface{})  {}
func (n *nopLogger) Errorw(string, ...interface{}) {}

//goland:noinspection SpellCheckingInspection
func (n *nopLogger) Panicw(string, ...interface{}) {}

//goland:noinspection SpellCheckingInspection
func (n *nopLogger) Fatalw(string, ...interface{}) {}

func (n *nopLogger) With(...interface{}) Logger {
	return n
}
func (n *nopLogger) WithOptions(...Option) Logger {
	return n
}
