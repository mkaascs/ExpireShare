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

func (fs *FileService) GetFileByAlias(ctx context.Context, command dto.GetFileCommand) (*dto.GetFileResult, error) {
	const fn = "services.FileService.GetFileByAlias"
	fs.log = slog.With(slog.String("fn", fn))

	fileInfo, err := fs.repo.GetFileByAlias(ctx, command.Alias)
	if err != nil {
		if errors.Is(err, repository.ErrAliasNotFound) {
			fs.log.Info("failed to get file info", sl.Error(err))
			return nil, ErrAliasNotFound
		}

		fs.log.Error("failed to get file info", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to get file info: %w", fn, err)
	}

	err = fs.checkPassword(fileInfo, command.Password)
	if err != nil {
		fs.log.Info("failed to check password", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to check password: %w", fn, err)
	}

	res := dto.GetFileResult{
		DownloadsLeft: fileInfo.DownloadsLeft,
		ExpiresIn:     time.Until(fileInfo.ExpiresAt),
	}

	return &res, nil
}
