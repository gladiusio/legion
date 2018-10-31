package logger

// Generic is the logger interface that legion uses, you can plug in
// your own logger as long as your logger implements this interface.
type Generic interface {
	Debug(msg string, keyvals ...interface{}) error
	Info(msg string, keyvals ...interface{}) error
	Error(msg string, keyvals ...interface{}) error

	With(keyvals ...interface{}) Generic
}
