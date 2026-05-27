package logger

import (
	"os"
	"github.com/rs/zerolog"
)

func NewLogger(level string) zerolog.Logger {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.DebugLevel
	}
	return zerolog.New(os.Stdout).
		With().
		Timestamp().
		Caller().
		Logger().
		Level(lvl)
}