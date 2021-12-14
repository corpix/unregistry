package middleware

import (
	"fmt"
	"io"

	"git.backbone/corpix/unregistry/pkg/log"

	glog "github.com/labstack/gommon/log"
)

type Logger struct {
	log.Logger
}

func (l Logger) Unwrap() log.Logger {
	return l.Logger
}

func (l Logger) logJSON(e *log.Event, v interface{}) {
	e.Interface("json", v)
}

func (l Logger) Debug(i ...interface{}) {
	l.Logger.Debug().Msg(fmt.Sprint(i...))
}

func (l Logger) Debugf(format string, i ...interface{}) {
	l.Logger.Debug().Msgf(format, i...)
}

func (l Logger) Debugj(j glog.JSON) {
	l.logJSON(l.Logger.Debug(), j)
}

func (l Logger) Info(i ...interface{}) {
	l.Logger.Info().Msg(fmt.Sprint(i...))
}

func (l Logger) Infof(format string, i ...interface{}) {
	l.Logger.Info().Msgf(format, i...)
}

func (l Logger) Infoj(j glog.JSON) {
	l.logJSON(l.Logger.Info(), j)
}

func (l Logger) Warn(i ...interface{}) {
	l.Logger.Warn().Msg(fmt.Sprint(i...))
}

func (l Logger) Warnf(format string, i ...interface{}) {
	l.Logger.Warn().Msgf(format, i...)
}

func (l Logger) Warnj(j glog.JSON) {
	l.logJSON(l.Logger.Warn(), j)
}

func (l Logger) Error(i ...interface{}) {
	l.Logger.Error().Msg(fmt.Sprint(i...))
}

func (l Logger) Errorf(format string, i ...interface{}) {
	l.Logger.Error().Msgf(format, i...)
}

func (l Logger) Errorj(j glog.JSON) {
	l.logJSON(l.Logger.Error(), j)
}

func (l Logger) Fatal(i ...interface{}) {
	l.Logger.Fatal().Msg(fmt.Sprint(i...))
}

func (l Logger) Fatalf(format string, i ...interface{}) {
	l.Logger.Fatal().Msgf(format, i...)
}

func (l Logger) Fatalj(j glog.JSON) {
	l.logJSON(l.Logger.Fatal(), j)
}

func (l Logger) Panic(i ...interface{}) {
	l.Logger.Panic().Msg(fmt.Sprint(i...))
}

func (l Logger) Panicf(format string, i ...interface{}) {
	l.Logger.Panic().Msgf(format, i...)
}

func (l Logger) Panicj(j glog.JSON) {
	l.logJSON(l.Logger.Panic(), j)
}

func (l Logger) Print(i ...interface{}) {
	l.Logger.Info().Msg(fmt.Sprint(i...))
}

func (l Logger) Printf(format string, i ...interface{}) {
	l.Logger.Info().Msgf(format, i...)
}

func (l Logger) Printj(j glog.JSON) {
	l.logJSON(l.Logger.Info(), j)
}

func (l Logger) Output() io.Writer {
	return l.Logger
}

// garbage methods

func (l *Logger) SetOutput(newOut io.Writer) {
	// not implemented
}

func (l Logger) Level() glog.Lvl {
	switch l.Logger.GetLevel() {
	case log.Debug:
		return glog.DEBUG
	case log.Trace:
		return glog.DEBUG
	case log.Info:
		return glog.INFO
	case log.Warn:
		return glog.WARN
	case log.Error:
		return glog.ERROR
	default:
		level := glog.Lvl(0)
		l.Logger.Warn().
			Str("level", l.Logger.GetLevel().String()).
			Uint8("default", uint8(level)).
			Msg("failed to map logger middleware log level, using default")
		return level
	}

}

func (l *Logger) SetLevel(level glog.Lvl) {
	// not implemented
}

func (l Logger) Prefix() string {
	return ""
}

func (l Logger) SetHeader(h string) {
	// not implemented
}

func (l *Logger) SetPrefix(prefix string) {
	// not implemented
}
