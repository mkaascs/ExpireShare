package mysql

import (
	"database/sql"
	"errors"
	"expire-share/internal/lib/log/sl"
	"fmt"
	"github.com/golang-migrate/migrate"
	"log/slog"
	"os"
)

type App struct {
	DB     *sql.DB
	logger *slog.Logger
}

func New(logger *slog.Logger, connectionString string) (*App, error) {
	const fn = "app.mysql.App.New"
	log := logger.With(slog.String("fn", fn))

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Error("failed to open db connection", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to open db connection: %w", fn, err)
	}

	return &App{
		DB:     db,
		logger: logger,
	}, nil
}

func (a *App) MustConnect() {
	if err := a.Connect(); err != nil {
		os.Exit(1)
	}
}

func (a *App) Connect() error {
	const fn = "app.mysql.App.Connect"
	log := a.logger.With(slog.String("fn", fn))

	if err := a.DB.Ping(); err != nil {
		log.Error("failed to ping db", sl.Error(err))
		return fmt.Errorf("%s: failed to ping db: %w", fn, err)
	}

	log.Info("connected to database successfully", slog.String("driver", "mysql"))
	return nil
}

func (a *App) Close() error {
	const fn = "app.mysql.App.Close"
	log := a.logger.With(slog.String("fn", fn))

	if err := a.DB.Close(); err != nil {
		log.Error("failed to close db", sl.Error(err))
		return fmt.Errorf("%s: failed to close db: %w", fn, err)
	}

	log.Info("closed database successfully", slog.String("driver", "mysql"))
	return nil
}

func MustMigrate(logger *slog.Logger, connectionString string) {
	if err := Migrate(logger, connectionString); err != nil {
		os.Exit(1)
	}
}

func Migrate(logger *slog.Logger, connectionString string) error {
	const fn = "app.mysql.App.Migrate"
	log := logger.With(slog.String("fn", fn))

	mgr, err := migrate.New("file://migrations", "mysql://"+connectionString)
	if err != nil {
		log.Error("failed to open migrations", sl.Error(err))
		return fmt.Errorf("%s: failed to open migrations: %w", fn, err)
	}

	defer func(mgr *migrate.Migrate) {
		if err, _ := mgr.Close(); err != nil {
			log.Error("failed to close migrator", sl.Error(err))
		}
	}(mgr)

	if err := mgr.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error("failed to up migrations", sl.Error(err))
		return fmt.Errorf("%s: failed to up migrations: %w", fn, err)
	}

	log.Info("migrated database successfully", slog.String("driver", "mysql"))
	return nil
}
