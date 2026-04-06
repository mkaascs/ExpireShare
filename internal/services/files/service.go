package files

import (
	"expire-share/internal/config"
	"expire-share/internal/domain/interfaces/repositories"
	"expire-share/internal/domain/interfaces/storage"
	"log/slog"
)

type Service struct {
	fileRepo    repositories.FileRepo
	fileStorage storage.File
	cfg         config.Config
	log         *slog.Logger
}

func New(fileRepo repositories.FileRepo, fileStorage storage.File, log *slog.Logger, cfg config.Config) *Service {
	return &Service{fileRepo: fileRepo,
		fileStorage: fileStorage,
		log:         log,
		cfg:         cfg}
}
