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
var _ Generic = (*ZeroLogger)(nil)

// ZeroLogger is a wrapper around zerolog so it can implement our interface
type ZeroLogger struct {
	Logger zerolog.Logger
}

// Debug calls the debug method of the registered logger
func (zl *ZeroLogger) Debug(msg string, keyvals ...interface{}) {
	if !validateKeyvals(keyvals...) {
		zl.Logger.Error().Str("debug_message", msg).
			Msg("Debug logger function keyvals are not valid, may be missing context...")
		return
	}

	addFields(zl.Logger.Debug(), keyvals...).Msg(msg)
}

// Info calls the info method of the registered logger
func (zl *ZeroLogger) Info(msg string, keyvals ...interface{}) {
	if !validateKeyvals(keyvals...) {
		zl.Logger.Error().Str("info_message", msg).
			Msg("Info logger function keyvals are not valid, may be missing context...")
		return
	}

	addFields(zl.Logger.Info(), keyvals...).Msg(msg)
}

// Warn calls the warn method of the registered logger
func (zl *ZeroLogger) Warn(msg string, keyvals ...interface{}) {
	if !validateKeyvals(keyvals...) {
		zl.Logger.Error().Str("warn_message", msg).
			Msg("Warn logger function keyvals are not valid, may be missing context...")
		return
	}

	addFields(zl.Logger.Warn(), keyvals...).Msg(msg)
}

// Error calls the error method of the registered logger
func (zl *ZeroLogger) Error(msg string, keyvals ...interface{}) {
	if !validateKeyvals(keyvals...) {
		zl.Logger.Error().Str("error_message", msg).
			Msg("Error logger function keyvals are not valid, may be missing context...")
		return
	}

	addFields(zl.Logger.Error(), keyvals...).Msg(msg)
}

// With calls the with method of the registered logger, and returns
// a logger with those fields attached by default
func (zl *ZeroLogger) With(keyvals ...interface{}) Generic {
	if !validateKeyvals(keyvals...) {
		zl.Logger.Error().
			Msg("Provided keyvals don't match correct pattern for With() method")
		return zl
	}

	l := zl.Logger.With()
	for i := 0; i < len(keyvals); i += 2 {
		l = l.Interface(keyvals[i].(string), keyvals[i+1])
	}

	return &ZeroLogger{Logger: l.Logger()}
}

func addFields(e *zerolog.Event, keyvals ...interface{}) *zerolog.Event {
	for i := 0; i < len(keyvals); i += 2 {
		e = e.Interface(keyvals[i].(string), keyvals[i+1])
	}

	return e
}

func validateKeyvals(keyvals ...interface{}) bool {
	if len(keyvals)%2 != 0 || len(keyvals) == 0 {
		return false
	}

	for i := 0; i < len(keyvals); i += 2 {
		if _, ok := keyvals[i].(string); !ok {
			return false
		}
	}

	return true
}
