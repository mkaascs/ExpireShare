package mysql

import "database/sql"

const (
	duplicateEntryErrCode = 1062
)

func stmtClose(stmt *sql.Stmt, err *error) {
	if *err != nil {
		_ = stmt.Close()
		return
	}

	*err = stmt.Close()
}
