package worker

import (
	"context"
	"expire-share/internal/config"
	"expire-share/internal/domain/interfaces/repositories"
	"expire-share/internal/lib/log/sl"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type FileWorker struct {
	Delay time.Duration
	cfg   config.Config
	repo  repositories.FileRepo
	log   *slog.Logger
}

func (fw *FileWorker) Start(ctx context.Context) {
	const fn = "services.worker.FileWorker.Start"
	fw.log = slog.With(slog.String("fn", fn))

	ticker := time.NewTicker(fw.Delay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fw.log.Info("file worker stopped")
			return

		case <-ticker.C:
			aliases, err := fw.repo.DeleteExpiredFiles(ctx)
			if err != nil {
				fw.log.Error("failed to delete expired files", sl.Error(err))
			}

			for _, alias := range aliases {
				folderPath := filepath.Join(fw.cfg.Path, alias)
				if err := os.RemoveAll(folderPath); err != nil {
					fw.log.Error("failed to delete file from storage", sl.Error(err))
				}
			}

			fw.log.Info(fmt.Sprintf("deleted %d expired files", len(aliases)))
		}
	}
}

func NewFileWorker(repo repositories.FileRepo, log *slog.Logger, cfg config.Config) *FileWorker {
	return &FileWorker{
		Delay: cfg.FileWorkerDelay,
		cfg:   cfg,
		repo:  repo,
		log:   log,
	}
}
