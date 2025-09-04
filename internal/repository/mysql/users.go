package mysql

import (
	"context"
	"database/sql"
	"errors"
	"expire-share/internal/domain/errors/repository"
	"expire-share/internal/domain/models"
	"fmt"
)

type UserRepo struct {
	Database *sql.DB
}

func (ur *UserRepo) GetUserById(ctx context.Context, userId int64) (_ models.User, err error) {
	const fn = "repository.mysql.UserRepo.GetUserById"

	stmt, err := ur.Database.PrepareContext(ctx, `SELECT login, password_hash FROM users WHERE id = ?`)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer stmtClose(stmt, &err)

	user := models.User{Id: userId}
	err = stmt.QueryRowContext(ctx, userId).Scan(
		&user.Login,
		&user.PasswordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, repository.ErrUserNotFound
		}

		return models.User{}, fmt.Errorf("%s: failed to query statement: %w", fn, err)
	}

	return user, nil
}

func (ur *UserRepo) GetUserByLogin(ctx context.Context, login string) (_ models.User, err error) {
	const fn = "repository.mysql.UserRepo.GetUserByLogin"

	stmt, err := ur.Database.PrepareContext(ctx, `SELECT id, password_hash FROM users WHERE login = ?`)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer stmtClose(stmt, &err)
	user := models.User{Login: login}
	err = stmt.QueryRowContext(ctx, login).Scan(
		&user.Id,
		&user.PasswordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, repository.ErrUserNotFound
		}

		return models.User{}, fmt.Errorf("%s: failed to query statement: %w", fn, err)
	}

	return user, nil
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{Database: db}
}
