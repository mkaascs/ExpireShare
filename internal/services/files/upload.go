package files

import (
	"context"
	"expire-share/internal/domain/dto/files/commands"
	"expire-share/internal/lib/alias"
	"expire-share/internal/lib/log/sl"
	"fmt"
	"io"
	"log/slog"
	"os"
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

	newAlias := alias.Gen(fs.cfg.AliasLength)

	fileDir := filepath.Join(fs.cfg.Path, newAlias)
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		log.Error("failed to create file dir", sl.Error(err))
		return "", fmt.Errorf("%s: failed to create file dir: %w", fn, err)
	}

	filePath := filepath.Join(fileDir, command.Filename)

	uploadedFile, err := os.Create(filePath)
	if err != nil {
		log.Error("failed to create file", sl.Error(err))
		return "", fmt.Errorf("%s: failed to create file: %w", fn, err)
	}

	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Error("failed to close file", sl.Error(err))
		}
	}(uploadedFile)

	if _, err := io.Copy(uploadedFile, command.File); err != nil {
		log.Error("failed to copy file", sl.Error(err))
		return "", fmt.Errorf("%s: failed to copy file: %w", fn, err)
	}

	var hashedBytes []byte
	if len(command.Password) > 0 {
		hashedBytes, err = bcrypt.GenerateFromPassword([]byte(command.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error("failed to hash password", sl.Error(err))
			return "", fmt.Errorf("%s: failed to hash password: %w", fn, err)
		}
	}

	addFileCommand := commands.AddFile{
		FilePath:     filepath.Join(newAlias, command.Filename),
		Alias:        newAlias,
		MaxDownloads: command.MaxDownloads,
		TTL:          command.TTL,
		PasswordHash: string(hashedBytes),
		UserID:       command.UserID,
	}

	_, err = fs.fileRepo.AddFile(ctx, addFileCommand)
	if err != nil {
		log.Error("failed to add file", sl.Error(err))
		return "", fmt.Errorf("%s: failed to add file: %w", fn, err)
	}

	return newAlias, nil
}
