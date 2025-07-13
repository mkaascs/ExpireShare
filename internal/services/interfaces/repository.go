package interfaces

import (
	"context"
	"expire-share/internal/domain"
	"expire-share/internal/services/dto"
)

type FileRepo interface {
	AddFile(ctx context.Context, command dto.AddFileCommand) (int64, error)
	GetFileByAlias(ctx context.Context, alias string) (domain.File, error)
	DeleteFile(ctx context.Context, alias string) error
}
