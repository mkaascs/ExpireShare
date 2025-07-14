package services

import (
	"errors"
	"expire-share/internal/config"
	"expire-share/internal/services/interfaces"
	"log/slog"
)

var (
	ErrFileSizeTooBig = errors.New("file size too big")
)

type FileService struct {
	repo interfaces.FileRepo
	cfg  config.Config
	log  *slog.Logger
}

func NewFileService(repo interfaces.FileRepo, log *slog.Logger, cfg config.Config) *FileService {
	return &FileService{repo: repo, log: log, cfg: cfg}
}
