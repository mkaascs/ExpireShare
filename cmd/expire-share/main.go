package main

import (
	"context"
	"errors"
	"expire-share/internal/config"
	"expire-share/internal/delivery/http/download"
	"expire-share/internal/delivery/http/file/get"
	"expire-share/internal/delivery/http/file/remove"
	"expire-share/internal/delivery/http/upload"
	myMiddleware "expire-share/internal/delivery/middlewares"
	pkgLog "expire-share/internal/lib/log"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/repository/mysql"
	"expire-share/internal/services"
	"expire-share/internal/services/worker"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"log"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	cfg := config.MustLoad()

	lg, err := pkgLog.New(cfg.Environment)
	if err != nil {
		log.Fatal("failed to initialize logger:", err)
	}

	lg.Info("starting expire share", slog.String("environment", cfg.Environment))

	repo, err := mysql.NewFileRepo(cfg.ConnectionString)
	if err != nil {
		lg.Error("failed to initialize repository:", sl.Error(err))
		os.Exit(1)
	}

	defer func() {
		err := repo.Database.Close()
		if err != nil {
			lg.Error("failed to close repository:", sl.Error(err))
		}
	}()

	lg.Info("repository was initialized successfully", slog.String("connection_string", cfg.ConnectionString))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fileWorker := worker.NewFileWorker(repo, lg, *cfg)
	go fileWorker.Start(ctx)

	lg.Info("file worker was started")

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(myMiddleware.NewLogger(lg))

	fileService := services.NewFileService(repo, lg, *cfg)
	router.Post("/upload", upload.New(fileService, lg, *cfg))
	router.Get("/download/{alias}", download.New(fileService, lg))
	router.Get("/file/{alias}", get.New(fileService, lg))
	router.Delete("/file/{alias}", remove.New(fileService, lg))

	lg.Info("starting expire share server", slog.String("address", cfg.HttpServer.Address))

	server := http.Server{
		Addr:         cfg.HttpServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		lg.Error("failed to start http server:", sl.Error(err))
	}

	lg.Error("server stopped")
}
