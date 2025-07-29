package interfaces

import (
	"context"
	"expire-share/internal/domain"
	"expire-share/internal/services/dto"
)

type FileService interface {
	UploadFile(ctx context.Context, command dto.UploadFileCommand) (string, error)
	DownloadFile(ctx context.Context, command dto.DownloadFileCommand) (*dto.DownloadFileResult, error)
	GetFileByAlias(ctx context.Context, command dto.GetFileCommand) (*dto.GetFileResult, error)
	DeleteFile(ctx context.Context, command dto.DeleteFileCommand) error
}

type UserService interface {
	Login(ctx context.Context, command dto.LoginCommand) (*domain.TokenPair, error)
}
