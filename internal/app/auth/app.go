package auth

import (
	"crypto/tls"
	"expire-share/internal/config"
	"expire-share/internal/lib/log/sl"
	"fmt"
	"log/slog"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	conn, err := grpc.NewClient(
		a.config.Addr,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))

	if err != nil {
		log.Error("failed to dial grpc connection", sl.Error(err))
		return fmt.Errorf("failed to dial grpc connection: %w", err)
	}

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

	return nil
}
