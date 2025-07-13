package services

import (
	"expire-share/internal/services/interfaces"
	"log/slog"
)

type FileService struct {
	repo interfaces.FileRepo
	log  *slog.Logger
}

func NewFileService(repo interfaces.FileRepo, log *slog.Logger) *FileService {
	return &FileService{repo: repo, log: log}
}
