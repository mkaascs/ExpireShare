package mysql

import (
	"database/sql"
	"errors"
	"expire-share/internal/domain"
	"expire-share/internal/repository"
	"expire-share/internal/services/dto"
	"fmt"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

const (
	duplicateEntryErrCode = 1062
)

type FileRepo struct {
	Database *sql.DB
}

func (r *FileRepo) AddFile(command dto.AddFileCommand) (_ int64, err error) {
	const fn = "storage.mysql.AddFile"

	stmt, err := r.Database.Prepare(`INSERT INTO files(file_path, alias, downloads_left, loaded_at, expires_at) VALUES(?, ?, ?, ?, ?)`)
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
			return 0, fmt.Errorf("%s: %w", fn, repository.ErrAliasExists)
		}

		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", fn, err)
	}

	return id, nil
}

func (r *FileRepo) GetFileByAlias(alias string) (domain.File, error) {
	const fn = "storage.mysql.GetFileByAlias"

	var file domain.File
	err := r.Database.QueryRow(`SELECT file_path, alias, downloads_left, loaded_at, expires_at FROM files WHERE alias = ?`, alias).Scan(
		&file.FilePath,
		&file.Alias,
		&file.DownloadsLeft,
		&file.LoadedAt,
		&file.ExpiresAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.File{}, repository.ErrAliasNotFound
		}

		return domain.File{}, fmt.Errorf("%s: %w", fn, err)
	}

	return file, nil
}

func (r *FileRepo) DeleteFile(alias string) (err error) {
	const fn = "storage.mysql.DeleteFile"

	stmt, err := r.Database.Prepare("DELETE FROM files WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	defer func(stmt *sql.Stmt) {
		if err != nil {
			_ = stmt.Close()
			return
		}

		err = stmt.Close()
	}(stmt)

	res, err := stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	if rowsAffected == 0 {
		return repository.ErrAliasNotFound
	}

	return nil
}

func NewFileRepo(connectionString string) (*FileRepo, error) {
	const fn = "storage.mysql.New"

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open database: %v", fn, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: failed to ping database: %v", fn, err)
	}

	return &FileRepo{Database: db}, nil
}
