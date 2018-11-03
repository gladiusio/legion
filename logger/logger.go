package logger

var internalLogger GenericLogger = NewZeroLogger()

// SetLogger sets the internal logger to the one provided
func SetLogger(newLogger GenericLogger) {
	internalLogger = newLogger
}

func GetLogger() GenericLogger {
	return internalLogger
}

// Debug calls the debug method of the registered logger
func Debug() GenericContext {
	return internalLogger.Debug()
}

//Info calls the info method of the registered logger
func Info() GenericContext {
	return internalLogger.Info()
}

// Warn calls the warn method of the registered logger
func Warn() GenericContext {
	return internalLogger.Warn()
}

// Error calls the error method of the registered logger
func Error() GenericContext {
	return internalLogger.Error()
}

// With calls the with method of the registered logger, and returns
// a logger with those fields attached by default
func With(gc GenericContext) GenericLogger {
	return internalLogger.With(gc)
}
