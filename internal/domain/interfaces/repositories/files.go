package repositories

import (
	"context"
	"expire-share/internal/domain/dto/files/commands"
	"expire-share/internal/domain/entities"
)

type FileRepo interface {
	AddFile(ctx context.Context, command commands.AddFile) (int64, error)
	GetFileByAlias(ctx context.Context, alias string) (entities.File, error)
	GetFilesByUserID(ctx context.Context, userID int64) ([]*entities.File, error)
	CountByUserID(ctx context.Context, userID int64) (int, error)
	DecrementDownloadsByAlias(ctx context.Context, alias string) (int16, error)
	DeleteFile(ctx context.Context, alias string) error
	DeleteExpiredFiles(ctx context.Context) ([]string, error)
}
