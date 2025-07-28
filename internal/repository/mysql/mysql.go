package mysql

import (
	"database/sql"
	"fmt"
)

const (
	duplicateEntryErrCode = 1062
)

func Connect(connectionString string) (*sql.DB, error) {
	const fn = "repository.mysql.Connect"

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open connection: %w", fn, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: failed to ping database: %w", fn, err)
	}

	return db, nil
}

func stmtClose(stmt *sql.Stmt, err *error) {
	if *err != nil {
		_ = stmt.Close()
		return
	}

	*err = stmt.Close()
}
