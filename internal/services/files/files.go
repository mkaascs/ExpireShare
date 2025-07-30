package files

import (
	"context"
	"errors"
	"expire-share/internal/config"
	"expire-share/internal/domain/errors/repository"
	"expire-share/internal/domain/errors/services/files"
	"expire-share/internal/domain/models"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/services/interfaces"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

type Service struct {
	fileRepo interfaces.FileRepo
	cfg      config.Config
	log      *slog.Logger
}

func (fs *Service) checkPasswordByAlias(ctx context.Context, alias string, password string) error {
	const fn = "services.file.Service.auth"
	fs.log = slog.With(slog.String("fn", fn))

	fileInfo, err := fs.fileRepo.GetFileByAlias(ctx, alias)
	if err != nil {
		if errors.Is(err, repository.ErrAliasNotFound) {
			fs.log.Info("failed to delete file info", sl.Error(err))
			return files.ErrAliasNotFound
		}

		fs.log.Error("failed to get file info", sl.Error(err))
		return err
	}

	return fs.checkPassword(fileInfo, password)
}

func (fs *Service) checkPassword(fileInfo models.File, password string) error {
	if fileInfo.PasswordHash != "" && password == "" {
		fs.log.Info("password is required for access")
		return files.ErrPasswordRequired
	}

	err := bcrypt.CompareHashAndPassword([]byte(fileInfo.PasswordHash), []byte(password))
	if err != nil && fileInfo.PasswordHash != "" {
		fs.log.Info("incorrect password", sl.Error(err))
		return files.ErrIncorrectPassword
	}

	return nil
}

func New(fileRepo interfaces.FileRepo, log *slog.Logger, cfg config.Config) *Service {
	return &Service{fileRepo: fileRepo,
		log: log,
		cfg: cfg}
}
