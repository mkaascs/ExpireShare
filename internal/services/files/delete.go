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
		if errors.Is(err, domainErrors.ErrFileNotFound) {
			log.Info(msg, sl.Error(err), slog.String("alias", command.Alias))
			return domainErrors.ErrFileNotFound
		}

		log.Error(msg, sl.Error(err), slog.String("alias", command.Alias))
		return fmt.Errorf("%s: %s: %w", fn, msg, err)
	}

	err = fs.checkAccess(fileInfo, command.UserID, command.Roles, command.Password)
	if err != nil {
		log.Info("access denied", sl.Error(err), slog.Int64("requesting_user_id", command.UserID), slog.String("alias", command.Alias))
		return fmt.Errorf("%s: access denied: %w", fn, err)
	}

	err = fs.fileRepo.DeleteFile(ctx, command.Alias)
	if err != nil {
		log.Error("failed to delete file info", sl.Error(err))
		return fmt.Errorf("%s: failed to delete file info: %w", fn, err)
	}

	if err := fs.fileStorage.Delete(command.Alias); err != nil {
		log.Error("failed to delete file from storage", sl.Error(err))
		return fmt.Errorf("%s: failed to delete file from storage: %w", fn, err)
	}

	return nil
}
