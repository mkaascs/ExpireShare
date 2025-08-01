package interfaces

import (
	"context"
	"expire-share/internal/domain/models"
	"expire-share/internal/services/dto/repository"
)

type FileRepo interface {
	AddFile(ctx context.Context, command repository.AddFileCommand) (int64, error)
	GetFileByAlias(ctx context.Context, alias string) (models.File, error)
	DecrementDownloadsByAlias(ctx context.Context, alias string) (int16, error)
	DeleteFile(ctx context.Context, alias string) error
	DeleteExpiredFiles(ctx context.Context) ([]string, error)
}

type TokenRepo interface {
	SaveToken(ctx context.Context, command repository.SaveTokenCommand) error
	GetToken(ctx context.Context, token string) (models.Token, error)
	ReplaceToken(ctx context.Context, userId int64, newTokenHash string) error
}

type UserRepo interface {
	GetUserById(ctx context.Context, userId int64) (models.User, error)
	GetUserByLogin(ctx context.Context, login string) (models.User, error)
}
