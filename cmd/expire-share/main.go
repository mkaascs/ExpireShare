package main

import (
	"expire-share/internal/config"
	pkgLog "expire-share/internal/lib/log"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/storage/mysql"
	"log"
	"log/slog"
	"os"
)

func main() {
	cfg := config.MustLoad()

	lg, err := pkgLog.New(cfg.Environment)
	if err != nil {
		log.Fatal("failed to initialize logger:", err)
	}

	lg.Info("starting expire share server", slog.String("environment", cfg.Environment))

	storage, err := mysql.New(cfg.ConnectionString)
	if err != nil {
		lg.Error("failed to initialize storage:", sl.Error(err))
		os.Exit(1)
	}

	defer func() {
		err := storage.Database.Close()
		if err != nil {
			lg.Error("failed to close storage:", sl.Error(err))
		}
	}()
}
