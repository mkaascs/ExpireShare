package interfaces

import (
	"context"
	"expire-share/internal/domain"
	"expire-share/internal/services/dto"
)

type FileRepo interface {
	AddFile(ctx context.Context, command dto.AddFileCommand) (int64, error)
	GetFileByAlias(ctx context.Context, alias string) (domain.File, error)
	DecrementDownloadsByAlias(ctx context.Context, alias string) (int16, error)
	DeleteFile(ctx context.Context, alias string) error
	DeleteExpiredFiles(ctx context.Context) ([]string, error)
}

type UserRepo interface {
	AddUser(ctx context.Context, command dto.AddUserCommand) (int64, error)
	GetUserById(ctx context.Context, id int64) (domain.User, error)
	GetUserAliasesById(ctx context.Context, id int64) ([]string, error)
}
