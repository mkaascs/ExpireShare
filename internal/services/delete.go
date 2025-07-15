package services

import (
	"context"
	"errors"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/repository"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func (fs *FileService) DeleteFile(ctx context.Context, alias string) error {
	const fn = "services.FileService.DeleteFile"
	fs.log = slog.With(slog.String("fn", fn))

	err := fs.repo.DeleteFile(ctx, alias)
	if err != nil {
		if errors.Is(err, repository.ErrAliasNotFound) {
			fs.log.Info("failed to delete file info", sl.Error(err))
			return ErrAliasNotFound
		}

		fs.log.Error("failed to delete file info", sl.Error(err))
		return fmt.Errorf("%s: failed to delete file info: %w", fn, err)
	}

	folderPath := filepath.Join(fs.cfg.Path, alias)
	if err := os.RemoveAll(folderPath); err != nil {
		fs.log.Error("failed to delete file from storage", sl.Error(err))
		return fmt.Errorf("%s: failed to delete file from storage: %w", fn, err)
	}

	return nil
}
