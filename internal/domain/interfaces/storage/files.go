package storage

import (
	"expire-share/internal/domain/dto/files/results"
	"io"
)

type File interface {
	Delete(alias string) error
	Download(alias string) (*results.DownloadFile, error)
	Upload(file io.Reader, alias string, filename string) error
}
