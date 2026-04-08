package local

import (
	"context"
	"expire-share/internal/config"
	"expire-share/internal/domain/dto/files/results"
	domainErrors "expire-share/internal/domain/entities/errors"
	"expire-share/internal/lib/log/sl"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type FileStorage struct {
	cfg config.Storage
	log *slog.Logger
}

func NewFileStorage(cfg config.Storage, log *slog.Logger) *FileStorage {
	return &FileStorage{cfg: cfg, log: log}
}

func (fs *FileStorage) Upload(ctx context.Context, file io.Reader, alias string, filename string) error {
	const fn = "storage.local.FileStorage.Upload"
	log := fs.log.With(slog.String("fn", fn))

	folderPath := filepath.Join(fs.cfg.Path, alias)

	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return fmt.Errorf("%s: create file folder failed: %w", fn, err)
	}

	filePath := filepath.Join(folderPath, filepath.Base(filename))
	destFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("%s: create file failed: %w", fn, err)
	}

	defer func(destFile *os.File) {
		if err := destFile.Close(); err != nil {
			log.Error("failed to close dest file", sl.Error(err))
		}
	}(destFile)

	if err := ctx.Err(); err != nil {
		return err
	}

	if _, err := io.Copy(destFile, file); err != nil {
		_ = destFile.Close()
		_ = os.RemoveAll(folderPath)
		return fmt.Errorf("%s: upload file failed: %w", fn, err)
	}

	return destFile.Sync()
}

func (fs *FileStorage) Download(_ context.Context, alias string) (*results.DownloadFile, error) {
	const fn = "storage.local.FileStorage.Download"

	folderPath := filepath.Join(fs.cfg.Path, alias)

	result, err := os.ReadDir(folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domainErrors.ErrFileNotFound
		}

		return nil, fmt.Errorf("%s: read dir failed: %w", fn, err)
	}

	if len(result) == 0 {
		return nil, domainErrors.ErrFileNotFound
	}

	filePath := filepath.Join(folderPath, result[0].Name())

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("%s: open file failed: %w", fn, err)
	}

	fileInfo, err := result[0].Info()
	if err != nil {
		return nil, fmt.Errorf("%s: get file info failed: %w", fn, err)
	}

	return &results.DownloadFile{
		File:     file,
		FileInfo: fileInfo,
		Close:    file.Close,
	}, nil
}

func (fs *FileStorage) Delete(_ context.Context, alias string) error {
	const fn = "storage.local.FileStorage.Delete"

	folderPath := filepath.Join(fs.cfg.Path, alias)
	if err := os.RemoveAll(folderPath); err != nil {
		if os.IsNotExist(err) {
			return domainErrors.ErrFileNotFound
		}

		return fmt.Errorf("%s: delete dir failed: %w", fn, err)
	}

	return nil
}
