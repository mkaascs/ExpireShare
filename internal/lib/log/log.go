package log

import (
	"expire-share/internal/config"
	"fmt"
	"log"
	"log/slog"
	"os"
)

func MustLoad(environment string) *slog.Logger {
	lg, err := Load(environment)
	if err != nil {
		log.Fatal(err)
	}

	return lg
}

func Load(environment string) (*slog.Logger, error) {
	var lg *slog.Logger

	switch environment {
	case config.EnvironmentLocal:
		lg = slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug}))
	case config.EnvironmentDev:
		lg = slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug}))
	case config.EnvironmentProd:
		lg = slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return nil, fmt.Errorf("unknown environment: %s", environment)
	}

	return lg, nil
}
