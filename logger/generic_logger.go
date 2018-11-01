package logger

// Generic is the logger interface that legion uses, you can plug in
// your own logger as long as your logger implements this interface.
type Generic interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})

	With(keyvals ...interface{}) Generic
}
