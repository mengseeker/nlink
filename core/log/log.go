package log

import (
	"os"

	"go.uber.org/zap"
)

type Logger = zap.SugaredLogger

func NewLogger() *Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.Development = false
	cfg.DisableStacktrace = true
	if os.Getenv("DEBUG") != "1" {
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	l, _ := cfg.Build(zap.AddCaller())
	return l.Sugar()
}
