package files

import (
	"context"
	"errors"
	"expire-share/internal/domain/dto/files/commands"
	domainErrors "expire-share/internal/domain/entities/errors"
	"expire-share/internal/lib/log/sl"
	"fmt"
	"log/slog"
)

func (fs *Service) DeleteFile(ctx context.Context, command commands.DeleteFile) error {
	const fn = "services.files.Service.DeleteFile"
	log := fs.log.With(slog.String("fn", fn))

	fileInfo, err := fs.fileRepo.GetFileByAlias(ctx, command.Alias)
	if err != nil {
		const msg = "failed to get file by alias"
		if errors.Is(err, domainErrors.ErrFileNotFound) || isCtxError(err) {
			log.Info(msg, sl.Error(err), slog.String("alias", command.Alias))
			return err
		}

		log.Error(msg, sl.Error(err), slog.String("alias", command.Alias))
		return fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	err = fs.checkAccess(*fileInfo, command.UserID, command.Roles, command.Password)
	if err != nil {
		log.Info("access denied", sl.Error(err), slog.Int64("requesting_user_id", command.UserID), slog.String("alias", command.Alias))
		return fmt.Errorf("%s: access denied: %w", fn, err)
	}

	tx, err := fs.fileRepo.BeginTx(ctx)
	if err != nil {
		log.Error("failed to begin tx", sl.Error(err))
		return fmt.Errorf("%s: failed to begin tx: %w", fn, err)
	}

	success := false
	defer func() {
		if !success {
			if err := tx.Rollback(); err != nil {
				log.Error("failed to rollback tx", sl.Error(err))
			}
		}
	}()

	err = fs.fileRepo.DeleteFileTx(ctx, tx, command.Alias)
	if err != nil {
		const msg = "failed to delete file info"
		if isCtxError(err) {
			log.Info(msg, sl.Error(err), slog.String("alias", command.Alias))
			return err
		}

		log.Error(msg, sl.Error(err), slog.String("alias", command.Alias))
		return fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	if err := fs.fileStorage.Delete(ctx, command.Alias); err != nil {
		const msg = "failed to delete file from storage"
		if isCtxError(err) {
			log.Info(msg, sl.Error(err), slog.String("alias", command.Alias))
			return err
		}

		log.Error(msg, sl.Error(err), slog.String("alias", command.Alias))
		return fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	if err := tx.Commit(); err != nil {
		log.Error("failed to commit tx", sl.Error(err))
		return fmt.Errorf("%s: failed to commit tx: %w", fn, err)
	}

	success = true
	return nil
}

func isCtxError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
