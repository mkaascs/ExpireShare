package files

import (
	"context"
	"expire-share/internal/domain/dto/files/commands"
	domainErrors "expire-share/internal/domain/entities/errors"
	"expire-share/internal/lib/alias"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/lib/sizes"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func (fs *Service) UploadFile(ctx context.Context, command commands.UploadFile) (string, error) {
	const fn = "services.file.Service.UploadFile"
	fs.log = slog.With(slog.String("fn", fn))

	if command.FileSize > fs.cfg.MaxFileSizeInBytes {
		fs.log.Info("file size too big",
			slog.String("file_size", sizes.ToFormattedString(command.FileSize)),
			slog.String("max_file_size", fs.cfg.MaxFileSize))

		return "", fmt.Errorf("%s: %w - max file size %s", fn, domainErrors.ErrFileSizeTooBig, fs.cfg.MaxFileSize)
	}

	newAlias := alias.Gen(fs.cfg.AliasLength)

	fileDir := filepath.Join(fs.cfg.Path, newAlias)
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		fs.log.Error("failed to create file dir", sl.Error(err))
		return "", fmt.Errorf("%s: failed to create file dir: %w", fn, err)
	}

	filePath := filepath.Join(fileDir, command.Filename)

	uploadedFile, err := os.Create(filePath)
	if err != nil {
		fs.log.Error("failed to create file", sl.Error(err))
		return "", fmt.Errorf("%s: failed to create file: %w", fn, err)
	}

	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			fs.log.Error("failed to close file", sl.Error(err))
		}

	}(uploadedFile)

	if _, err := io.Copy(uploadedFile, command.File); err != nil {
		fs.log.Error("failed to copy file", sl.Error(err))
		return "", fmt.Errorf("%s: failed to copy file: %w", fn, err)
	}

	var hashedBytes []byte
	if len(command.Password) > 0 {
		hashedBytes, err = bcrypt.GenerateFromPassword([]byte(command.Password), bcrypt.DefaultCost)
		if err != nil {
			fs.log.Error("failed to hash password", sl.Error(err))
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
		fs.log.Error("failed to add file", sl.Error(err))
		return "", fmt.Errorf("%s: failed to add file: %w", fn, err)
	}

	return newAlias, nil
}
