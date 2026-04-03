package files

import (
	"context"
	"errors"
	"expire-share/internal/domain/dto/files/commands"
	"expire-share/internal/domain/dto/files/results"
	domainErrors "expire-share/internal/domain/entities/errors"
	"expire-share/internal/lib/log/sl"
	"fmt"
	"log/slog"
	"time"
)

func (fs *Service) GetFileByAlias(ctx context.Context, command commands.GetFile) (*results.GetFile, error) {
	const fn = "services.file.Service.GetFileByAlias"
	fs.log = slog.With(slog.String("fn", fn))

	fileInfo, err := fs.fileRepo.GetFileByAlias(ctx, command.Alias)
	if err != nil {
		if errors.Is(err, domainErrors.ErrAliasNotFound) {
			fs.log.Info("failed to get file info", sl.Error(err))
			return nil, domainErrors.ErrAliasNotFound
		}

		fs.log.Error("failed to get file info", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to get file info: %w", fn, err)
	}

	err = fs.checkPassword(fileInfo, command.Password)
	if err != nil {
		fs.log.Info("failed to check password", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to check password: %w", fn, err)
	}

	res := results.GetFile{
		DownloadsLeft: fileInfo.DownloadsLeft,
		ExpiresIn:     time.Until(fileInfo.ExpiresAt),
	}

	return &res, nil
}
