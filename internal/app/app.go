package app

import (
	"context"
	"expire-share/internal/app/http"
	"expire-share/internal/app/mysql"
	"expire-share/internal/config"
	"expire-share/internal/delivery/http/api/files/get"
	"expire-share/internal/delivery/http/api/files/remove"
	"expire-share/internal/delivery/http/api/upload"
	"expire-share/internal/delivery/http/download"
	myMiddleware "expire-share/internal/delivery/middlewares"
	repo "expire-share/internal/infrastructure/mysql"
	"expire-share/internal/services/files"
	"expire-share/internal/services/worker"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
)

type App struct {
	HTTP   *http.App
	MySql  *mysql.App
	config config.Config
	logger *slog.Logger
}

func New(config config.Config, logger *slog.Logger) *App {
	httpApp := http.New(logger, config.HttpServer)

	mysql.MustMigrate(logger, config.DbConnectionString)
	mysqlApp, _ := mysql.New(logger, config.DbConnectionString)

	return &App{
		HTTP:   httpApp,
		MySql:  mysqlApp,
		config: config,
		logger: logger,
	}
}

func (a *App) MustMountMiddlewares() {
	a.HTTP.Router.Use(middleware.RequestID)
	a.HTTP.Router.Use(middleware.RealIP)
	a.HTTP.Router.Use(middleware.Recoverer)
	a.HTTP.Router.Use(middleware.URLFormat)
	a.HTTP.Router.Use(myMiddleware.NewLogger(a.logger))
}

func (a *App) MustMountHandlers() {
	fileRepo := repo.NewFileRepo(a.MySql.DB)

	fileService := files.New(fileRepo, a.logger, a.config)

	if a.config.Env == config.EnvLocal {
		a.HTTP.Router.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
		))
	}

	a.HTTP.Router.Get("/download/{alias}", download.New(fileService, a.logger))

	a.HTTP.Router.Route("/api", func(r chi.Router) {
		r.Post("/upload", upload.New(fileService, a.logger, a.config))

		r.Route("/file", func(r chi.Router) {
			r.Get("/{alias}", get.New(fileService, a.logger))
			r.Delete("/{alias}", remove.New(fileService, a.logger))
		})
	})
}

func (a *App) StartFileWorker(ctx context.Context) {
	fileRepo := repo.NewFileRepo(a.MySql.DB)

	fileWorker := worker.NewFileWorker(fileRepo, a.logger, a.config)
	fileWorker.Start(ctx)
}
