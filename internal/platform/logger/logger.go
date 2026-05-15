package logger

import (
	"log/slog"
	"os"
)

var log *slog.Logger

func Init(isProduction bool) {
	if isProduction {
		log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	} else {
		log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
}

func Get() *slog.Logger {
	if log == nil {
		return slog.Default()
	}
	return log
}
