package interfaces

import (
	"context"
	"expire-share/internal/services/dto"
)

type FileService interface {
	UploadFile(ctx context.Context, command dto.UploadFileCommand) (string, error)
}
