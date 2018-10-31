package logger

// NewZeroLogger returns a new ZeroLogger
func NewZeroLogger() *ZeroLogger { return &ZeroLogger{} }

var _ Generic = (*ZeroLogger)(nil)

// ZeroLogger is a wrapper around zerolog so it can implement our interface
type ZeroLogger struct{}

// Debug calls the debug method of the registered logger
func (zl *ZeroLogger) Debug(msg string, keyvals ...interface{}) error {
	return nil
}

//Info calls the info method of the registered logger
func (zl *ZeroLogger) Info(msg string, keyvals ...interface{}) error {
	return nil

}

// Error calls the error method of the registered logger
func (zl *ZeroLogger) Error(msg string, keyvals ...interface{}) error {
	return nil

}

// With calls the with method of the registered logger, and returns
// a logger with those fields attached by default
func (zl *ZeroLogger) With(keyvals ...interface{}) Generic {
	return nil
}
