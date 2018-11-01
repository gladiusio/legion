package logger

var internalLogger Generic = NewZeroLogger()

// SetLogger sets the internal logger to the one provided
func SetLogger(newLogger Generic) {
	internalLogger = newLogger
}

// Debug calls the debug method of the registered logger
func Debug(msg string, keyvals ...interface{}) {
	internalLogger.Debug(msg, keyvals)
}

//Info calls the info method of the registered logger
func Info(msg string, keyvals ...interface{}) {
	internalLogger.Info(msg, keyvals)
}

// Error calls the error method of the registered logger
func Error(msg string, keyvals ...interface{}) {
	internalLogger.Error(msg, keyvals)
}

// With calls the with method of the registered logger, and returns
// a logger with those fields attached by default
func With(keyvals ...interface{}) Generic {
	return internalLogger.With(keyvals)
}
