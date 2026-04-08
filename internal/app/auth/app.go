package auth

import (
	"expire-share/internal/config"
	"expire-share/internal/lib/log/sl"
	"fmt"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"os"

	"google.golang.org/grpc"
)

type App struct {
	GRPCConn *grpc.ClientConn
	logger   *slog.Logger
	config   config.AuthService
}

func New(logger *slog.Logger, config config.AuthService) *App {
	return &App{
		logger: logger,
		config: config,
	}
}

func (a *App) MustConnect() {
	if err := a.Connect(); err != nil {
		os.Exit(1)
	}
}

func (a *App) Connect() error {
	const fn = "app.auth.App.Connect"
	log := a.logger.With(slog.String("fn", fn))

	conn, err := grpc.NewClient(
		a.config.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Error("failed to dial grpc connection", sl.Error(err))
		return fmt.Errorf("failed to dial grpc connection: %w", err)
	}

	log.Info("connected to grpc server successfully", slog.String("addr", a.config.Addr))

	a.GRPCConn = conn
	return nil
}

func (a *App) Close() error {
	const fn = "app.auth.App.Close"
	log := a.logger.With(slog.String("fn", fn))

	if err := a.GRPCConn.Close(); err != nil {
		log.Error("failed to close grpc connection", sl.Error(err))
		return fmt.Errorf("failed to close grpc connection: %w", err)
	}

	log.Info("grpc connection closed successfully", slog.String("addr", a.config.Addr))

	return nil
}
