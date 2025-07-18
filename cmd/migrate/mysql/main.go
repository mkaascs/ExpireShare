package main

import (
	"database/sql"
	"expire-share/internal/config"
	pkgLog "expire-share/internal/lib/log"
	"expire-share/internal/lib/log/sl"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	mysqlMigrate "github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	"log"
	"os"
)

func main() {
	migrateFunc := map[string]func(*migrate.Migrate) error{
		"up": func(mgr *migrate.Migrate) error {
			return mgr.Up()
		},
		"down": func(mgr *migrate.Migrate) error {
			return mgr.Down()
		},
	}

	cfg := config.MustLoad("")
	lg, err := pkgLog.New(cfg.Environment)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) < 2 {
		log.Fatal("usage: migrate [up/down]")
	}

	cmd := os.Args[1]

	db, err := sql.Open("mysql", cfg.ConnectionString)
	if err != nil {
		lg.Error("failed to open database", sl.Error(err))
		os.Exit(1)
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			lg.Error("failed to close database", sl.Error(err))
		}
	}(db)

	if err := db.Ping(); err != nil {
		lg.Error("failed to ping database", sl.Error(err))
		os.Exit(1)
	}

	driver, err := mysqlMigrate.WithInstance(db, &mysqlMigrate.Config{})
	if err != nil {
		lg.Error("failed to create database driver", sl.Error(err))
		os.Exit(1)
	}

	mgr, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"mysql",
		driver)

	if err != nil {
		lg.Error("failed to create migration instance", sl.Error(err))
		os.Exit(1)
	}

	fn, ok := migrateFunc[cmd]
	if !ok {
		lg.Error(fmt.Sprintf("unknown command: %s", cmd))
		os.Exit(1)
	}

	if err := fn(mgr); err != nil {
		lg.Error("failed to migrate table", sl.Error(err))
		os.Exit(1)
	}

	lg.Info(fmt.Sprintf("migrate %s finished successfully", cmd))
}
