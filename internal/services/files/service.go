package files

import (
	"expire-share/internal/config"
	"expire-share/internal/domain/interfaces/repositories"
	"log/slog"
)

type Service struct {
	fileRepo repositories.FileRepo
	cfg      config.Config
	log      *slog.Logger
}

func New(fileRepo repositories.FileRepo, log *slog.Logger, cfg config.Config) *Service {
	return &Service{fileRepo: fileRepo,
		log: log,
		cfg: cfg}
}
