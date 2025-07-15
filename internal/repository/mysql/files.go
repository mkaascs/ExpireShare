package mysql

import (
	"context"
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

func (fr *FileRepo) AddFile(ctx context.Context, command dto.AddFileCommand) (_ int64, err error) {
	const fn = "repository.mysql.AddFile"

	stmt, err := fr.Database.PrepareContext(ctx, `INSERT INTO files(file_path, alias, downloads_left, loaded_at, expires_at) VALUES(?, ?, ?, ?, ?)`)
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
	res, err := stmt.ExecContext(ctx,
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

func (fr *FileRepo) GetFileByAlias(ctx context.Context, alias string) (domain.File, error) {
	const fn = "repository.mysql.GetFileByAlias"

	var file domain.File
	err := fr.Database.QueryRowContext(ctx, `SELECT file_path, alias, downloads_left, loaded_at, expires_at FROM files WHERE alias = ? AND expires_at > NOW()`, alias).Scan(
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

func (fr *FileRepo) DecrementDownloadsByAlias(ctx context.Context, alias string) (int16, error) {
	const fn = "repository.mysql.DecrementDownloadsByAlias"

	var downloadsLeft int16
	err := fr.Database.QueryRowContext(ctx, `SELECT downloads_left FROM files WHERE alias = ? AND expires_at > NOW()`, alias).Scan(&downloadsLeft)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	if downloadsLeft == 0 {
		return 0, repository.ErrNoDownloadsLeft
	}

	downloadsLeft--
	res, err := fr.Database.ExecContext(ctx, `UPDATE files SET downloads_left = ? WHERE alias = ? AND expires_at > NOW()`, downloadsLeft, alias)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	if rowsAffected == 0 {
		return 0, repository.ErrAliasNotFound
	}

	return downloadsLeft, nil
}

func (fr *FileRepo) DeleteFile(ctx context.Context, alias string) (err error) {
	const fn = "repository.mysql.DeleteFile"

	stmt, err := fr.Database.PrepareContext(ctx, "DELETE FROM files WHERE alias = ? AND expires_at > NOW()")
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

	res, err := stmt.ExecContext(ctx, alias)
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
	const fn = "repository.mysql.NewFileRepo"

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open database: %v", fn, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: failed to ping database: %v", fn, err)
	}

	return &FileRepo{Database: db}, nil
}
