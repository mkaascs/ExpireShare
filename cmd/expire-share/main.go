package main

import (
	"expire-share/internal/config"
	pkgLog "expire-share/internal/pkg/log"
	"log"
	"log/slog"
)

func main() {
	cfg := config.MustLoad()

	lg, err := pkgLog.New(cfg.Environment)
	if err != nil {
		log.Fatal("failed to initialize logger:", err)
	}

	lg.Info("starting expire share server", slog.String("environment", cfg.Environment))
}
