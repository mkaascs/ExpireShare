package main

// @title Expire Share API
// @version 1.0
// @description A self-destructing file-sharing service with TTL and download limits

import (
	"context"
	"expire-share/internal/app"
	"expire-share/internal/config"
	myLog "expire-share/internal/lib/log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()
	logger := myLog.MustLoad(cfg.Env)

	logger.Info("application expire-share is starting", slog.String("env", cfg.Env))

	application := app.New(*cfg, logger)

	application.MySql.MustConnect()
	application.Auth.MustConnect()

	application.MustMountMiddlewares()
	application.MustMountHandlers()

	go application.HTTP.MustRun()
	go application.StartFileWorker(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	_ = application.MySql.Close()
	_ = application.Auth.Close()
	_ = application.HTTP.Shutdown(ctx)

	logger.Info("application expire-share stopped")
}
