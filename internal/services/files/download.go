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
)

func (fs *Service) DownloadFile(ctx context.Context, command commands.DownloadFile) (*results.DownloadFile, error) {
	const fn = "services.files.Service.DownloadFile"
	log := fs.log.With(slog.String("fn", fn))

	fileInfo, err := fs.fileRepo.GetFileByAlias(ctx, command.Alias)
	if err != nil {
		const msg = "failed to get file by alias"
		if errors.Is(err, domainErrors.ErrFileNotFound) {
			log.Info(msg, sl.Error(err), slog.String("alias", command.Alias))
			return nil, domainErrors.ErrFileNotFound
		}

		log.Error(msg, sl.Error(err), slog.String("alias", command.Alias))
		return nil, fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	err = fs.checkPassword(fileInfo, command.Password)
	if err != nil {
		log.Info("access denied", sl.Error(err), slog.String("alias", command.Alias))
		return nil, fmt.Errorf("%s: access denied: %w", fn, err)
	}

	downloadsLeft, err := fs.fileRepo.DecrementDownloadsByAlias(ctx, command.Alias)
	if err != nil {
		log.Error("failed to decrement downloads left", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to decrement downloads left: %w", fn, err)
	}

	result, err := fs.fileStorage.Download(command.Alias)
	if err != nil {
		log.Error("failed to download file from storage", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to download file from storage: %w", fn, err)
	}

	if downloadsLeft > 0 {
		return result, nil
	}

	return &results.DownloadFile{
		File:     result.File,
		FileInfo: result.FileInfo,
		Close: func() error {
			if err := result.Close(); err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}

			if err := fs.fileRepo.DeleteFile(ctx, command.Alias); err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}

			if err := fs.fileStorage.Delete(command.Alias); err != nil {
				return fmt.Errorf("%s: %w", fn, err)
			}

			return nil
		},
	}, nil
}
