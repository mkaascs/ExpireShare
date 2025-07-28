package mysql

import (
	"context"
	"database/sql"
	"expire-share/internal/services/dto"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type TokenRepo struct {
	Database *sql.DB
}

func (tr *TokenRepo) SaveToken(ctx context.Context, command dto.SaveTokenCommand) (_ int64, err error) {
	const fn = "repository.mysql.TokenRepo.SaveToken"

	stmt, err := tr.Database.PrepareContext(ctx, `INSERT INTO tokens(user_id, refresh_token) VALUES(?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer stmtClose(stmt, &err)
	res, err := stmt.ExecContext(ctx, command.UserId, command.RefreshTokenHash)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to exec statement: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", fn, err)
	}

	return id, nil
}

func (tr *TokenRepo) CheckToken(ctx context.Context, refreshToken string) (bool, error) {
	const fn = "repository.mysql.TokenRepo.CheckToken"

	stmt, err := tr.Database.PrepareContext(ctx, `SELECT EXISTS(SELECT 1 FROM tokens WHERE refresh_token = ? AND expires_at > NOW())`)
	if err != nil {
		return false, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return false, fmt.Errorf("%s: failed to hash refresh token: %w", fn, err)
	}

	var exists bool
	err = stmt.QueryRowContext(ctx, string(hashedBytes)).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%s: failed to check token existence: %w", fn, err)
	}

	return exists, nil
}

func NewTokenRepo(db *sql.DB) *TokenRepo {
	return &TokenRepo{Database: db}
}
