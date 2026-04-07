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

	result, err := fs.fileStorage.Download(command.Alias)
	if err != nil {
		log.Error("failed to download file from storage", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to download file from storage: %w", fn, err)
	}

	tx, err := fs.fileRepo.BeginTx(ctx)
	if err != nil {
		if err := result.Close(); err != nil {
			log.Error("failed to close file", sl.Error(err))
		}

		log.Error("failed to begin tx", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to begin tx: %w", fn, err)
	}

	success := false
	defer func() {
		if !success {
			if err := tx.Rollback(); err != nil {
				log.Error("failed to rollback tx", sl.Error(err))
			}

			if err := result.Close(); err != nil {
				log.Error("failed to close file", sl.Error(err))
			}
		}
	}()

	downloadsLeft, err := fs.fileRepo.DecrementDownloadsByAliasTx(ctx, tx, command.Alias)
	if err != nil {
		log.Error("failed to decrement downloads left", sl.Error(err), slog.String("alias", command.Alias))
		return nil, fmt.Errorf("%s: failed to decrement downloads left: %w", fn, err)
	}

	if downloadsLeft > 0 {
		if err := tx.Commit(); err != nil {
			log.Error("failed to commit tx", sl.Error(err))
			return nil, fmt.Errorf("%s: failed to commit tx: %w", fn, err)
		}

		success = true
		return result, nil
	}

	if err := fs.fileRepo.DeleteFileTx(ctx, tx, command.Alias); err != nil {
		log.Error("failed to delete file from repo", sl.Error(err), slog.String("alias", command.Alias))
		return nil, fmt.Errorf("%s: failed to delete from DB: %w", fn, err)
	}

	if err := fs.fileStorage.Delete(command.Alias); err != nil {
		log.Error("failed to delete file from storage", sl.Error(err), slog.String("alias", command.Alias))
		return nil, fmt.Errorf("%s: failed to delete from storage: %w", fn, err)
	}

	if err := tx.Commit(); err != nil {
		log.Error("failed to commit tx", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to commit delete: %w", fn, err)
	}

	success = true
	return result, nil
}
