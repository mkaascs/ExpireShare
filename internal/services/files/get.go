package files

import (
	"context"
	"errors"
	"expire-share/internal/domain/errors/repository"
	"expire-share/internal/domain/errors/services/files"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/services/dto/commands"
	"expire-share/internal/services/dto/results"
	"fmt"
	"log/slog"
	"time"
)

func (fs *Service) GetFileByAlias(ctx context.Context, command commands.GetFileCommand) (*results.GetFileResult, error) {
	const fn = "services.file.Service.GetFileByAlias"
	fs.log = slog.With(slog.String("fn", fn))

	fileInfo, err := fs.fileRepo.GetFileByAlias(ctx, command.Alias)
	if err != nil {
		if errors.Is(err, repository.ErrAliasNotFound) {
			fs.log.Info("failed to get file info", sl.Error(err))
			return nil, files.ErrAliasNotFound
		}

		fs.log.Error("failed to get file info", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to get file info: %w", fn, err)
	}

	err = fs.checkPassword(fileInfo, command.Password)
	if err != nil {
		fs.log.Info("failed to check password", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to check password: %w", fn, err)
	}

	res := results.GetFileResult{
		DownloadsLeft: fileInfo.DownloadsLeft,
		ExpiresIn:     time.Until(fileInfo.ExpiresAt),
	}

	return &res, nil
}
