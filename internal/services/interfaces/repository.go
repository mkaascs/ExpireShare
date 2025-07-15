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
