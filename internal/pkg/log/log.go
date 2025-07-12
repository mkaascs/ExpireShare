package log

import (
	"fmt"
	"log/slog"
	"os"
)

const (
	environmentLocal = "local"
	environmentDev   = "dev"
	environmentProd  = "prod"
)

func New(environment string) (*slog.Logger, error) {
	var lg *slog.Logger

	switch environment {
	case environmentLocal:
		lg = slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug}))
	case environmentDev:
		lg = slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug}))
	case environmentProd:
		lg = slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return nil, fmt.Errorf("unknown environment: %s", environment)
	}

	return lg, nil
}
