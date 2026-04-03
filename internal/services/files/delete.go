package files

import (
	"context"
	"errors"
	"expire-share/internal/domain/dto/files/commands"
	domainErrors "expire-share/internal/domain/entities/errors"
	"expire-share/internal/lib/log/sl"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func (fs *Service) DeleteFile(ctx context.Context, command commands.DeleteFile) error {
	const fn = "services.files.Service.DeleteFile"
	fs.log = slog.With(slog.String("fn", fn))

	err := fs.checkPasswordByAlias(ctx, command.Alias, command.Password)
	if err != nil {
		fs.log.Info("failed to check password", sl.Error(err))
		return fmt.Errorf("%s: failed to check password: %w", fn, err)
	}

	err = fs.fileRepo.DeleteFile(ctx, command.Alias)
	if err != nil {
		if errors.Is(err, domainErrors.ErrAliasNotFound) {
			fs.log.Info("failed to delete file info", sl.Error(err))
			return domainErrors.ErrAliasNotFound
		}

		fs.log.Error("failed to delete file info", sl.Error(err))
		return fmt.Errorf("%s: failed to delete file info: %w", fn, err)
	}

	folderPath := filepath.Join(fs.cfg.Path, command.Alias)
	if err := os.RemoveAll(folderPath); err != nil {
		fs.log.Error("failed to delete file from storage", sl.Error(err))
		return fmt.Errorf("%s: failed to delete file from storage: %w", fn, err)
	}

	return nil
}
