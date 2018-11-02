package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// NewZeroLogger returns a new ZeroLogger
func NewZeroLogger() *ZeroLogger {
	l := zerolog.New(os.Stderr).With().Timestamp().Logger()
	l = l.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	return &ZeroLogger{Logger: l}
}

// Compile time assertion that our logger meets the intferface specifications
var _ GenericLogger = (*ZeroLogger)(nil)
var _ GenericContext = (*ZeroLogContext)(nil)

type field struct {
	key string
	val interface{}
}

type ZeroLogContext struct {
	event  *zerolog.Event
	fields []field
}

func (zlc *ZeroLogContext) Field(key string, val interface{}) GenericContext {
	zlc.event = zlc.event.Interface(key, val)
	zlc.fields = append(zlc.fields, field{key, val})
	return zlc
}

func (zlc *ZeroLogContext) Log(msg string) {
	zlc.event.Msg(msg)
}

// ZeroLogger is a wrapper around zerolog so it can implement our interface
type ZeroLogger struct {
	Logger zerolog.Logger
}

// Debug calls the debug method of the registered logger
func (zl *ZeroLogger) Debug() GenericContext {
	return &ZeroLogContext{event: zl.Logger.Debug(), fields: make([]field, 0)}
}

// Info calls the info method of the registered logger
func (zl *ZeroLogger) Info() GenericContext {
	return &ZeroLogContext{event: zl.Logger.Info(), fields: make([]field, 0)}
}

// Warn calls the warn method of the registered logger
func (zl *ZeroLogger) Warn() GenericContext {
	return &ZeroLogContext{event: zl.Logger.Warn(), fields: make([]field, 0)}
}

// Error calls the error method of the registered logger
func (zl *ZeroLogger) Error() GenericContext {
	return &ZeroLogContext{event: zl.Logger.Error(), fields: make([]field, 0)}
}

// With appends the context to a new logger and returns it
func (zl *ZeroLogger) With(gc GenericContext) GenericLogger {
	ctx := zl.Logger.With()
	for _, field := range gc.(*ZeroLogContext).fields {
		ctx = ctx.Interface(field.key, field.val)
	}
	return &ZeroLogger{Logger: ctx.Logger()}
}
