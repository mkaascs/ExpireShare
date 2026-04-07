package storage

import (
	"context"
	"expire-share/internal/domain/dto/files/results"
	"io"
)

type File interface {
	Delete(ctx context.Context, alias string) error
	Download(ctx context.Context, alias string) (*results.DownloadFile, error)
	Upload(ctx context.Context, file io.Reader, alias string, filename string) error
}
