package log

import (
	"fmt"

	"golang.org/x/exp/slog"
)

type Logger struct {
	slog.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Logger: *slog.Default(),
	}
}

func (l *Logger) Debugf(format string, vals ...any) {
	l.Logger.Debug(fmt.Sprintf(format, vals...))
}

func (l *Logger) Infof(format string, vals ...any) {
	l.Logger.Info(fmt.Sprintf(format, vals...))
}

func (l *Logger) Warnf(format string, vals ...any) {
	l.Logger.Warn(fmt.Sprintf(format, vals...))
}

func (l *Logger) Errorf(format string, vals ...any) {
	l.Logger.Error(fmt.Sprintf(format, vals...))
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: *l.Logger.With(args...),
	}
}
