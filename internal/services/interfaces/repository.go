package interfaces

import (
	"expire-share/internal/domain"
	"expire-share/internal/services/dto"
)

type FileRepo interface {
	AddFile(command dto.AddFileCommand) (int64, error)
	GetFileByAlias(alias string) (domain.File, error)
	DeleteFile(alias string) error
}
