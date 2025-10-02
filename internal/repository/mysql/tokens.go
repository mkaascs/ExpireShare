package mysql

import (
	"context"
	"database/sql"
	"errors"
	repositoryErr "expire-share/internal/domain/errors/repository"
	"expire-share/internal/domain/models"
	"expire-share/internal/services/dto/repository"
	"fmt"
	"github.com/go-sql-driver/mysql"
)

type TokenRepo struct {
	Database *sql.DB
}

func (tr *TokenRepo) SaveToken(ctx context.Context, command repository.SaveTokenCommand) (_ int64, err error) {
	const fn = "repository.mysql.TokenRepo.SaveToken"

	stmt, err := tr.Database.PrepareContext(ctx, `INSERT INTO tokens(user_id, refresh_token_hash, expires_at) VALUES (?,?,?) ON DUPLICATE KEY UPDATE refresh_token_hash = VALUES(refresh_token_hash), expires_at = VALUES(expires_at)`)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer stmtClose(stmt, &err)

	res, err := stmt.ExecContext(ctx, command.UserId, command.RefreshTokenHash, command.ExpiresAt)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == duplicateEntryErrCode {
			return 0, repositoryErr.ErrTokenExists
		}

		return 0, fmt.Errorf("%s: failed to save token: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get token id: %w", fn, err)
	}

	return id, nil
}

func (tr *TokenRepo) GetToken(ctx context.Context, refreshTokenHash string) (_ models.Token, err error) {
	const fn = "repository.mysql.TokenRepo.GetToken"

	stmt, err := tr.Database.PrepareContext(ctx, `SELECT user_id, refresh_token_hash, expires_at FROM tokens WHERE refresh_token_hash = ?`)
	if err != nil {
		return models.Token{}, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer stmtClose(stmt, &err)

	var token models.Token
	err = stmt.QueryRowContext(ctx, refreshTokenHash).Scan(
		&token.UserId,
		&token.Hash,
		&token.ExpiresAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Token{}, repositoryErr.ErrTokenNotFound
		}

		return models.Token{}, fmt.Errorf("%s: failed to query row: %w", fn, err)
	}

	return token, nil
}

func (tr *TokenRepo) ReplaceToken(ctx context.Context, userId int64, newTokenHash string) error {
	const fn = "repository.mysql.TokenRepo.ReplaceToken"

	stmt, err := tr.Database.PrepareContext(ctx, `UPDATE tokens SET refresh_token_hash = ? WHERE user_id = ?`)
	if err != nil {
		return fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer stmtClose(stmt, &err)

	_, err = stmt.ExecContext(ctx, newTokenHash, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repositoryErr.ErrTokenNotFound
		}

		return fmt.Errorf("%s: failed to update row: %w", fn, err)
	}

	return nil
}

func NewTokenRepo(db *sql.DB) *TokenRepo {
	return &TokenRepo{Database: db}
}
