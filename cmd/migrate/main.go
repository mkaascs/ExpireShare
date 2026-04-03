package main

import (
	"database/sql"
	"expire-share/internal/config"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/source/file"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate [up/down]")
	}

	cfg := config.MustLoad()

	cmd := os.Args[1]

	db, err := sql.Open("mysql", cfg.DbConnectionString)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	defer func(db *sql.DB) {
		if err := db.Close(); err != nil {
			log.Fatalf("failed to close database: %v", err)
		}
	}(db)

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	mgr, err := migrate.New("file://migrations", "mysql://"+cfg.DbConnectionString)
	if err != nil {
		log.Fatalf("failed to init migrator: %v", err)
	}

	operations := map[string]func() error{
		"up":   mgr.Up,
		"down": mgr.Down,
	}

	fn, ok := operations[cmd]
	if !ok {
		log.Fatalf(fmt.Sprintf("unknown command: %s", cmd))
	}

	if err := fn(); err != nil {
		log.Fatalf("failed to migrate table: %v", err)
	}

	fmt.Println("migration complete")
}
