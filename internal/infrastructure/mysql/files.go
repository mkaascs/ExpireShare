package mysql

import (
	"context"
	"database/sql"
	"errors"
	"expire-share/internal/domain/dto/files/commands"
	"expire-share/internal/domain/entities"
	domainErrors "expire-share/internal/domain/entities/errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

const (
	duplicateEntryErrCode = 1062
)

type FileRepo struct {
	DB  *sql.DB
	log *slog.Logger
}

func NewFileRepo(db *sql.DB, log *slog.Logger) *FileRepo {
	return &FileRepo{DB: db, log: log}
}

func (fr *FileRepo) AddFile(ctx context.Context, command commands.AddFile) (int64, error) {
	const fn = "repository.mysql.FileRepo.AddFile"
	log := fr.log.With(slog.String("fn", fn))

	stmt, err := fr.DB.PrepareContext(ctx, `INSERT INTO files(file_path, alias, downloads_left, loaded_at, expires_at, password_hash, user_id) VALUES(?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			log.Error("%s: failed to close stmt: %w", fn, err)
		}
	}(stmt)

	currentTime := time.Now()
	res, err := stmt.ExecContext(ctx,
		command.FilePath,
		command.Alias,
		command.MaxDownloads,
		currentTime,
		currentTime.Add(command.TTL),
		command.PasswordHash,
		command.UserID)

	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == duplicateEntryErrCode {
			return 0, domainErrors.ErrAliasTaken
		}

		return 0, fmt.Errorf("%s: failed to exec statement: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", fn, err)
	}

	return id, nil
}

func (fr *FileRepo) GetFileByAlias(ctx context.Context, alias string) (entities.File, error) {
	const fn = "repository.mysql.FileRepo.GetFileByAlias"
	log := fr.log.With(slog.String("fn", fn))

	stmt, err := fr.DB.PrepareContext(ctx, `SELECT file_path, alias, downloads_left, loaded_at, expires_at, password_hash FROM files WHERE alias = ? AND expires_at > NOW()`)
	if err != nil {
		return entities.File{}, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			log.Error("%s: failed to close stmt: %w", fn, err)
		}
	}(stmt)

	var file entities.File
	err = stmt.QueryRowContext(ctx, alias).Scan(
		&file.FilePath,
		&file.Alias,
		&file.DownloadsLeft,
		&file.LoadedAt,
		&file.ExpiresAt,
		&file.PasswordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.File{}, domainErrors.ErrFileNotFound
		}

		return entities.File{}, fmt.Errorf("%s: failed to query statement: %w", fn, err)
	}

	return file, nil
}

func (fr *FileRepo) DecrementDownloadsByAlias(ctx context.Context, alias string) (int16, error) {
	const fn = "repository.mysql.FileRepo.DecrementDownloadsByAlias"
	log := fr.log.With(slog.String("fn", fn))

	tx, err := fr.DB.BeginTx(ctx, nil)
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

	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			log.Error("%s: failed to close stmt: %w", fn, err)
		}
	}(selectStmt)

	var downloadsLeft int16
	err = selectStmt.QueryRowContext(ctx, alias).Scan(&downloadsLeft)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to exec statement: %w", fn, err)
	}

	if downloadsLeft == 0 {
		return 0, domainErrors.ErrNoDownloadsLeft
	}

	downloadsLeft--
	updateStmt, err := tx.PrepareContext(ctx, `UPDATE files SET downloads_left = ? WHERE alias = ? AND expires_at > NOW()`)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			log.Error("%s: failed to close stmt: %w", fn, err)
		}
	}(updateStmt)

	res, err := updateStmt.ExecContext(ctx, downloadsLeft, alias)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to exec statement: %w", fn, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to affect rows: %w", fn, err)
	}

	if rowsAffected == 0 {
		return 0, domainErrors.ErrFileNotFound
	}

	return downloadsLeft, nil
}

func (fr *FileRepo) DeleteFile(ctx context.Context, alias string) error {
	const fn = "repository.mysql.FileRepo.DeleteFile"
	log := fr.log.With(slog.String("fn", fn))

	stmt, err := fr.DB.PrepareContext(ctx, "DELETE FROM files WHERE alias = ? AND expires_at > NOW()")
	if err != nil {
		return fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			log.Error("%s: failed to close stmt: %w", fn, err)
		}
	}(stmt)

	res, err := stmt.ExecContext(ctx, alias)
	if err != nil {
		return fmt.Errorf("%s: failed to exec statement: %w", fn, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to affect rows: %w", fn, err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrFileNotFound
	}

	return nil
}

func (fr *FileRepo) DeleteExpiredFiles(ctx context.Context) ([]string, error) {
	const fn = "repository.mysql.FileRepo.DeleteExpiredFiles"
	log := fr.log.With(slog.String("fn", fn))

	selectStmt, err := fr.DB.PrepareContext(ctx, `SELECT alias FROM files WHERE expires_at < NOW() FOR UPDATE`)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			log.Error("%s: failed to close stmt: %w", fn, err)
		}
	}(selectStmt)

	rows, err := selectStmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to exec statement: %w", fn, err)
	}

	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			log.Error("%s: failed to close rows: %w", fn, err)
		}
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

	updateStmt, err := fr.DB.PrepareContext(ctx, `DELETE FROM files WHERE expires_at < NOW()`)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			log.Error("%s: failed to close stmt: %w", fn, err)
		}
	}(updateStmt)

	_, err = updateStmt.ExecContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to exec statement: %w", fn, err)
	}

	return aliases, nil
}
