package services

import (
	"context"
	"errors"
	"expire-share/internal/config"
	"expire-share/internal/lib/alias"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/lib/sizes"
	"expire-share/internal/services/dto"
	"expire-share/internal/services/interfaces"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

var (
	ErrFileSizeTooBig = errors.New("file size too big")
)

type FileService struct {
	repo interfaces.FileRepo
	cfg  config.Config
	log  *slog.Logger
}

func (fs *FileService) UploadFile(ctx context.Context, command dto.UploadFileCommand) (_ string, err error) {
	const fn = "services.FileService.UploadFile"
	fs.log = fs.log.With(slog.String("fn", fn))

	if command.FileSize > fs.cfg.MaxFileSizeInBytes {
		fs.log.Info("file size too big",
			slog.String("file_size", sizes.ToFormattedString(command.FileSize)),
			slog.String("max_file_size", fs.cfg.MaxFileSize))
		return "", fmt.Errorf("%s: %w - max file size %s", fn, ErrFileSizeTooBig, fs.cfg.MaxFileSize)
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
		if err != nil {
			_ = file.Close()
			return
		}

		err = file.Close()
	}(uploadedFile)

	if _, err := io.Copy(uploadedFile, command.File); err != nil {
		fs.log.Error("failed to copy file", sl.Error(err))
		return "", fmt.Errorf("%s: failed to copy file: %w", fn, err)
	}

	addFileCommand := dto.AddFileCommand{
		FilePath:     filepath.Join(newAlias, command.Filename),
		Alias:        newAlias,
		MaxDownloads: command.MaxDownloads,
		TTL:          command.TTL,
	}

	_, err = fs.repo.AddFile(ctx, addFileCommand)
	if err != nil {
		fs.log.Error("failed to add file", sl.Error(err))
		return "", fmt.Errorf("%s: failed to add file: %w", fn, err)
	}

	return newAlias, nil
}

func NewFileService(repo interfaces.FileRepo, log *slog.Logger, cfg config.Config) *FileService {
	return &FileService{repo: repo, log: log, cfg: cfg}
}
