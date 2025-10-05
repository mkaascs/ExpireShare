package main

// @title Expire Share API
// @version 1.0
// @description A self-destructing file-sharing service with TTL and download limits

import (
	"context"
	"database/sql"
	"errors"
	"expire-share/internal/config"
	"expire-share/internal/crypto"
	"expire-share/internal/delivery/http/api/files/get"
	"expire-share/internal/delivery/http/api/files/remove"
	"expire-share/internal/delivery/http/api/upload"
	"expire-share/internal/delivery/http/auth/login"
	"expire-share/internal/delivery/http/auth/register"
	"expire-share/internal/delivery/http/auth/token/refresh"
	"expire-share/internal/delivery/http/download"
	myMiddleware "expire-share/internal/delivery/middlewares"
	myLog "expire-share/internal/lib/log"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/repository/mysql"
	"expire-share/internal/services/auth"
	"expire-share/internal/services/files"
	"expire-share/internal/services/worker"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
	"os"
)

type App struct {
	router chi.Router

	cfg *config.Config
	lg  *slog.Logger
	db  *sql.DB

	fileService *files.Service
	authService *auth.Service
}

func (a *App) Run() error {
	server := http.Server{
		Addr:         a.cfg.HttpServer.Address,
		Handler:      a.router,
		ReadTimeout:  a.cfg.Timeout,
		WriteTimeout: a.cfg.Timeout,
		IdleTimeout:  a.cfg.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		a.lg.Error("failed to start http server:", sl.Error(err))
		return err
	}

	return nil
}

func (a *App) Close() error {
	if err := a.db.Close(); err != nil {
		a.lg.Error("failed to close database connection:", sl.Error(err))
		return err
	}

	return nil
}

func (a *App) MountHandlers() {
	if a.cfg.Environment == config.EnvironmentLocal {
		a.router.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
		))
	}

	a.router.Get("/download/{alias}", download.New(a.fileService, a.lg))

	a.router.Route("/api", func(r chi.Router) {
		r.Post("/upload", upload.New(a.fileService, a.lg, *a.cfg))

		r.Route("/file", func(r chi.Router) {
			r.Get("/{alias}", get.New(a.fileService, a.lg))
			r.Delete("/{alias}", remove.New(a.fileService, a.lg))
		})
	})

	a.router.Route("/auth", func(r chi.Router) {
		r.With(myMiddleware.NewBodyParser[login.Request](a.lg)).
			With(myMiddleware.NewValidator[login.Request](a.lg)).
			Post("/login", login.New(a.authService, a.lg))

		r.With(myMiddleware.NewBodyParser[register.Request](a.lg)).
			With(myMiddleware.NewValidator[register.Request](a.lg)).
			Post("/register", register.New(a.authService, a.lg))

		r.With(myMiddleware.NewBodyParser[refresh.Request](a.lg)).
			With(myMiddleware.NewValidator[refresh.Request](a.lg)).
			Post("/token/refresh", refresh.New(a.authService, a.lg))
	})
}

func (a *App) MountMiddlewares() {
	a.router.Use(middleware.RequestID)
	a.router.Use(middleware.RealIP)
	a.router.Use(middleware.Recoverer)
	a.router.Use(middleware.URLFormat)
	a.router.Use(myMiddleware.NewLogger(a.lg))
}

func (a *App) StartFileWorker() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	fileWorker := worker.NewFileWorker(mysql.NewFileRepo(a.db), a.lg, *a.cfg)
	go fileWorker.Start(ctx)

	return cancel
}

func NewApp(envPath string) (*App, error) {
	app := &App{
		cfg: config.MustLoad(envPath),
	}

	app.lg = myLog.MustLoad(app.cfg.Environment)

	var err error
	app.db, err = mysql.Connect(app.cfg.ConnectionString)
	if err != nil {
		app.lg.Error("failed to connect to database:", sl.Error(err))
		return nil, err
	}

	fileRepo := mysql.NewFileRepo(app.db)
	userRepo := mysql.NewUserRepo(app.db)
	tokenRepo := mysql.NewTokenRepo(app.db)

	rsaKeyConfig, err := crypto.NewRsaKey(envPath)
	if err != nil {
		app.lg.Error("failed to load rsa key:", sl.Error(err))
		return nil, err
	}

	hmacConfig, err := crypto.NewHmacConfig(envPath)
	if err != nil {
		app.lg.Error("failed to load hmac config:", sl.Error(err))
		return nil, err
	}

	app.fileService = files.New(fileRepo, app.lg, *app.cfg)
	app.authService = auth.New(tokenRepo, userRepo, *app.cfg, app.lg, auth.Secrets{
		PrivateKey: rsaKeyConfig.GetPrivateKey(),
		HmacSecret: hmacConfig.GetHmacSecret()})

	app.router = chi.NewRouter()

	return app, nil
}

func main() {
	envPath := ""
	if len(os.Args) > 2 {
		envPath = os.Args[1]
	}

	app, err := NewApp(envPath)
	if err != nil {
		os.Exit(1)
	}

	app.lg.Info("expire share app is initialized", slog.String("enviroment", app.cfg.Environment))
	defer func(app *App) {
		_ = app.Close()
	}(app)

	cancel := app.StartFileWorker()
	defer cancel()

	app.lg.Info("file worker was started")

	app.MountMiddlewares()
	app.MountHandlers()

	app.lg.Info("starting expire share server", slog.String("address", app.cfg.HttpServer.Address))

	_ = app.Run()
	app.lg.Error("server stopped")
}
