package interfaces

import (
	"context"
	"expire-share/internal/services/dto"
)

type FileService interface {
	UploadFile(ctx context.Context, command dto.UploadFileCommand) (string, error)
	DownloadFile(ctx context.Context, alias string) (*dto.DownloadFileResult, error)
}
