package mysql

import (
	"database/sql"
	"errors"
	"expire-share/internal/storage"
	"fmt"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

const (
	duplicateEntryErrCode = 1062
)

type Storage struct {
	Database *sql.DB
}

func (s *Storage) Upload(command UploadFileCommand) (_ int64, err error) {
	const fn = "storage.mysql.Upload"

	stmt, err := s.Database.Prepare(`INSERT INTO files(file_path, alias, downloads_left, loaded_at, expires_at) VALUES(?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	defer func(stmt *sql.Stmt) {
		if err != nil {
			_ = stmt.Close()
			return
		}

		err = stmt.Close()
	}(stmt)

	currentTime := time.Now()
	res, err := stmt.Exec(
		command.FilePath,
		command.Alias,
		command.MaxDownloads,
		currentTime,
		currentTime.Add(command.TTL))

	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == duplicateEntryErrCode {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrAliasExists)
		}

		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", fn, err)
	}

	return id, nil
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

	return &Storage{Database: db}, nil
}
