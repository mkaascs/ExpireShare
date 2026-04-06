package files

import (
	"context"
	"expire-share/internal/domain/dto/files/commands"
	"expire-share/internal/lib/alias"
	"expire-share/internal/lib/log/sl"
	"fmt"
	"log/slog"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

func (fs *Service) UploadFile(ctx context.Context, command commands.UploadFile) (string, error) {
	const fn = "services.file.Service.UploadFile"
	log := fs.log.With(slog.String("fn", fn))

	filesCount, err := fs.fileRepo.CountByUserID(ctx, command.UserID)
	if err != nil {
		log.Info("failed to count files by user id", sl.Error(err), slog.Int64("user_id", command.UserID))
		return "", fmt.Errorf("%s: failed to count files by user id: %w", fn, err)
	}

	err = fs.checkUploadQuote(filesCount, command.FileSize, command.Roles)
	if err != nil {
		log.Info("access denied", sl.Error(err), slog.Int64("user_id", command.UserID))
		return "", fmt.Errorf("%s: failed to upload quote: %w", fn, err)
	}

	genAlias := alias.Gen(fs.cfg.AliasLength)

	// TODO: make nil if empty
	var hashedBytes []byte
	if len(command.Password) > 0 {
		hashedBytes, err = bcrypt.GenerateFromPassword([]byte(command.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error("failed to hash password", sl.Error(err))
			return "", fmt.Errorf("%s: failed to hash password: %w", fn, err)
		}
	}

	// TODO: remove filepath
	_, err = fs.fileRepo.AddFile(ctx, commands.AddFile{
		FilePath:     filepath.Join(genAlias, command.Filename),
		Alias:        genAlias,
		MaxDownloads: command.MaxDownloads,
		TTL:          command.TTL,
		PasswordHash: string(hashedBytes),
		UserID:       command.UserID,
	})

	if err != nil {
		log.Error("failed to add file", sl.Error(err))
		return "", fmt.Errorf("%s: failed to add file: %w", fn, err)
	}

	if err := fs.fileStorage.Upload(command.File, genAlias, command.Filename); err != nil {
		log.Error("failed to upload file to storage", sl.Error(err))
		return "", fmt.Errorf("%s: failed to upload file to storage: %w", fn, err)
	}

	return genAlias, nil
}
