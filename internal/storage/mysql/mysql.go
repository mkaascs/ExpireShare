package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	mysqlMigrate "github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
)

type Storage struct {
	Database *sql.DB
}

func New(connectionString string) (*Storage, error) {
	const fn = "storage.mysql.New"

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open database: %v", fn, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: failed to ping database: %v", fn, err)
	}

	driver, err := mysqlMigrate.WithInstance(db, &mysqlMigrate.Config{})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create database driver: %v", fn, err)
	}

	mgr, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"mysql",
		driver)

	if err != nil {
		return nil, fmt.Errorf("%s: failed to create migration instance: %v", fn, err)
	}

	if err := mgr.Up(); err != nil {
		return nil, fmt.Errorf("%s: failed to run migrations: %v", fn, err)
	}

	return &Storage{Database: db}, nil
}
