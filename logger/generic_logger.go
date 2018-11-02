package logger

// GenericLogger is the logger interface that legion uses, you can plug in
// your own logger as long as your logger implements this interface.
type GenericLogger interface {
	// Base log types
	Debug() GenericContext
	Info() GenericContext
	Warn() GenericContext
	Error() GenericContext

	// Add context like logger.With(NewContext().Field("test", "val"))
	With(ctx GenericContext) GenericLogger
}

// GenericContext provides a way to add fields to a log event
type GenericContext interface {
	Field(key string, val interface{}) GenericContext

	// Actually log the built up log line with the message
	Log(msg string)
}
