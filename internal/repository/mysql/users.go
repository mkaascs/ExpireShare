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
)

type UserRepo struct {
	Database *sql.DB
}

func (ur *UserRepo) AddUser(ctx context.Context, command dto.AddUserCommand) (_ int64, err error) {
	const fn = "repository.UserRepo.AddUser"

	stmt, err := ur.Database.Prepare(`INSERT INTO users(ip) VALUES(?)`)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer stmtClose(stmt, &err)
	res, err := stmt.ExecContext(ctx, command.IP)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == duplicateEntryErrCode {
			return 0, fmt.Errorf("%s: user already exists: %w", fn, err)
		}

		return 0, fmt.Errorf("%s: failed to exec statement: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert ID: %w", fn, err)
	}

	return id, nil
}

func (ur *UserRepo) GetUserById(ctx context.Context, id int64) (_ domain.User, err error) {
	const fn = "repository.UserRepo.GetUserById"

	stmt, err := ur.Database.Prepare(`SELECT ip FROM users WHERE id = ?`)
	if err != nil {
		return domain.User{}, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	defer stmtClose(stmt, &err)

	var user domain.User
	err = stmt.QueryRowContext(ctx, id).Scan(
		&user.IP)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, repository.ErrUserIdNotFound
		}

		return domain.User{}, fmt.Errorf("%s: failed to query statement: %w", fn, err)
	}

	return user, nil
}

func NewUserRepo(connectionString string) (*UserRepo, error) {
	const fn = "repository.mysql.NewUserRepo"

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open database: %w", fn, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: failed to ping database: %w", fn, err)
	}

	return &UserRepo{Database: db}, nil
}
