package files

import (
	"context"
	"errors"
	"expire-share/internal/config"
	"expire-share/internal/domain"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/repository"
	"expire-share/internal/services/interfaces"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

var (
	ErrFileSizeTooBig    = errors.New("file size too big")
	ErrAliasNotFound     = errors.New("file with current alias does not exist")
	ErrPasswordRequired  = errors.New("password required for access")
	ErrIncorrectPassword = errors.New("incorrect password")
)

type Service struct {
	fileRepo interfaces.FileRepo
	userRepo interfaces.UserRepo
	cfg      config.Config
	log      *slog.Logger
}

func (fs *Service) checkPasswordByAlias(ctx context.Context, alias string, password string) error {
	const fn = "services.FileService.auth"
	fs.log = slog.With(slog.String("fn", fn))

	fileInfo, err := fs.fileRepo.GetFileByAlias(ctx, alias)
	if err != nil {
		if errors.Is(err, repository.ErrAliasNotFound) {
			fs.log.Info("failed to delete file info", sl.Error(err))
			return ErrAliasNotFound
		}

		fs.log.Error("failed to get file info", sl.Error(err))
		return err
	}

	return fs.checkPassword(fileInfo, password)
}

func (fs *Service) checkPassword(fileInfo domain.File, password string) error {
	if fileInfo.PasswordHash != "" && password == "" {
		fs.log.Info("password is required for access")
		return ErrPasswordRequired
	}

	err := bcrypt.CompareHashAndPassword([]byte(fileInfo.PasswordHash), []byte(password))
	if err != nil && fileInfo.PasswordHash != "" {
		fs.log.Info("incorrect password", sl.Error(err))
		return ErrIncorrectPassword
	}

	return nil
}

func NewFileService(fileRepo interfaces.FileRepo, userRepo interfaces.UserRepo, log *slog.Logger, cfg config.Config) *Service {
	return &Service{fileRepo: fileRepo,
		userRepo: userRepo,
		log:      log,
		cfg:      cfg}
}
