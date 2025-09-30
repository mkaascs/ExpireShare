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

type UserRepo struct {
	Database *sql.DB
}

func (ur *UserRepo) AddUser(ctx context.Context, command repository.AddUserCommand) (int64, error) {
	const fn = "repository.mysql.UserRepo.AddUser"

	stmt, err := ur.Database.PrepareContext(ctx, `INSERT INTO users(login,password_hash) VALUES(?,?)`)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer stmtClose(stmt, &err)

	res, err := stmt.ExecContext(ctx, command.Login, command.PasswordHash)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == duplicateEntryErrCode {
			return 0, repositoryErr.ErrUserAlreadyExists
		}

		return 0, fmt.Errorf("%s: failed to add user: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get user id: %w", fn, err)
	}

	return id, nil
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
			return models.User{}, repositoryErr.ErrUserNotFound
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
			return models.User{}, repositoryErr.ErrUserNotFound
		}

		return models.User{}, fmt.Errorf("%s: failed to query statement: %w", fn, err)
	}

	return user, nil
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{Database: db}
}
