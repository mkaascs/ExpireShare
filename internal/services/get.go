package services

import (
	"context"
	"errors"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/repository"
	"expire-share/internal/services/dto"
	"fmt"
	"log/slog"
	"time"
)

func (fs *FileService) GetFileByAlias(ctx context.Context, alias string) (*dto.GetFileResult, error) {
	const fn = "services.FileService.GetFileByAlias"
	fs.log = slog.With(slog.String("fn", fn))

	file, err := fs.repo.GetFileByAlias(ctx, alias)
	if err != nil {
		if errors.Is(err, repository.ErrAliasNotFound) {
			fs.log.Info("failed to get file info", sl.Error(err))
			return nil, ErrAliasNotFound
		}

		fs.log.Error("failed to get file info", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to get file info: %w", fn, err)
	}

	res := dto.GetFileResult{
		DownloadsLeft: file.DownloadsLeft,
		ExpiresIn:     time.Until(file.ExpiresAt),
	}

	return &res, nil
}
