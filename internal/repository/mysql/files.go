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

type FileRepo struct {
	Database *sql.DB
}

func (fr *FileRepo) AddFile(ctx context.Context, command dto.AddFileCommand) (_ int64, err error) {
	const fn = "repository.mysql.AddFile"

	stmt, err := fr.Database.PrepareContext(ctx, `INSERT INTO files(file_path, alias, downloads_left, loaded_at, expires_at, password_hash) VALUES(?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer stmtClose(stmt, &err)

	currentTime := time.Now()
	res, err := stmt.ExecContext(ctx,
		command.FilePath,
		command.Alias,
		command.MaxDownloads,
		currentTime,
		currentTime.Add(command.TTL),
		command.PasswordHash)

	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == duplicateEntryErrCode {
			return 0, fmt.Errorf("%s: failed to exec statement: %w", fn, repository.ErrAliasExists)
		}

		return 0, fmt.Errorf("%s: failed to exec statement: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", fn, err)
	}

	return id, nil
}

func (fr *FileRepo) GetFileByAlias(ctx context.Context, alias string) (domain.File, error) {
	const fn = "repository.mysql.GetFileByAlias"

	stmt, err := fr.Database.PrepareContext(ctx, `SELECT file_path, alias, downloads_left, loaded_at, expires_at, password_hash FROM files WHERE alias = ? AND expires_at > NOW()`)
	if err != nil {
		return domain.File{}, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer stmtClose(stmt, &err)

	var file domain.File
	err = stmt.QueryRowContext(ctx, alias).Scan(
		&file.FilePath,
		&file.Alias,
		&file.DownloadsLeft,
		&file.LoadedAt,
		&file.ExpiresAt,
		&file.PasswordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.File{}, repository.ErrAliasNotFound
		}

		return domain.File{}, fmt.Errorf("%s: failed to query statement: %w", fn, err)
	}

	return file, nil
}

func (fr *FileRepo) DecrementDownloadsByAlias(ctx context.Context, alias string) (int16, error) {
	const fn = "repository.mysql.DecrementDownloadsByAlias"

	tx, err := fr.Database.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	defer func(tx *sql.Tx) {
		if err != nil {
			_ = tx.Rollback()
			return
		}

		err = tx.Commit()
	}(tx)

	selectStmt, err := tx.PrepareContext(ctx, `SELECT downloads_left FROM files WHERE alias = ? AND expires_at > NOW()`)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer stmtClose(selectStmt, &err)

	var downloadsLeft int16
	err = selectStmt.QueryRowContext(ctx, alias).Scan(&downloadsLeft)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to exec statement: %w", fn, err)
	}

	if downloadsLeft == 0 {
		return 0, repository.ErrNoDownloadsLeft
	}

	downloadsLeft--
	updateStmt, err := tx.PrepareContext(ctx, `UPDATE files SET downloads_left = ? WHERE alias = ? AND expires_at > NOW()`)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer stmtClose(updateStmt, &err)

	res, err := updateStmt.ExecContext(ctx, downloadsLeft, alias)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to exec statement: %w", fn, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to affect rows: %w", fn, err)
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
		return fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer stmtClose(stmt, &err)

	res, err := stmt.ExecContext(ctx, alias)
	if err != nil {
		return fmt.Errorf("%s: failed to exec statement: %w", fn, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to affect rows: %w", fn, err)
	}

	if rowsAffected == 0 {
		return repository.ErrAliasNotFound
	}

	return nil
}

func (fr *FileRepo) DeleteExpiredFiles(ctx context.Context) (_ []string, err error) {
	const fn = "repository.mysql.DeleteExpiredFiles"
	tx, err := fr.Database.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	defer func(tx *sql.Tx) {
		if err != nil {
			_ = tx.Rollback()
			return
		}

		err = tx.Commit()
	}(tx)

	selectStmt, err := tx.PrepareContext(ctx, `SELECT alias FROM files WHERE expires_at < NOW() FOR UPDATE`)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer stmtClose(selectStmt, &err)

	rows, err := selectStmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to exec statement: %w", fn, err)
	}

	defer func(rows *sql.Rows) {
		if err != nil {
			_ = rows.Close()
		}

		err = rows.Close()
	}(rows)

	var aliases []string
	for rows.Next() {
		var alias string
		if err := rows.Scan(&alias); err != nil {
			return nil, fmt.Errorf("%s: failed to scan alias: %w", fn, err)
		}

		aliases = append(aliases, alias)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	updateStmt, err := tx.PrepareContext(ctx, `DELETE FROM files WHERE expires_at < NOW()`)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer stmtClose(updateStmt, &err)

	_, err = updateStmt.ExecContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to exec statement: %w", fn, err)
	}

	return aliases, nil
}

func NewFileRepo(db *sql.DB) *FileRepo {
	return &FileRepo{Database: db}
}
