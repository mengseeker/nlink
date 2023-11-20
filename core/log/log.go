package log

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/exp/slog"
)

var Out io.Writer = os.Stdout

type Logger struct {
	slog.Logger
}

func NewLogger() *Logger {
	opt := slog.HandlerOptions{
		// AddSource: true,
		// Level: slog.LevelDebug,
	}
	if os.Getenv("DEBUG") == "true" {
		opt.Level = slog.LevelDebug
	}
	l := slog.New(slog.NewTextHandler(Out, &opt))
	return &Logger{
		Logger: *l,
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
