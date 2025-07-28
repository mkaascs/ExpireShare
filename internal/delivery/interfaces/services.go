package interfaces

import (
	"context"
	"expire-share/internal/services/dto"
)

type FileService interface {
	UploadFile(ctx context.Context, command dto.UploadFileCommand) (string, error)
	DownloadFile(ctx context.Context, command dto.DownloadFileCommand) (*dto.DownloadFileResult, error)
	GetFileByAlias(ctx context.Context, command dto.GetFileCommand) (*dto.GetFileResult, error)
	DeleteFile(ctx context.Context, command dto.DeleteFileCommand) error
}
