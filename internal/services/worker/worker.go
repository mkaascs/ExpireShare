package worker

import (
	"context"
	"expire-share/internal/config"
	"expire-share/internal/domain/interfaces/repositories"
	"expire-share/internal/domain/interfaces/storage"
	"expire-share/internal/lib/log/sl"
	"fmt"
	"log/slog"
	"time"
)

type FileWorker struct {
	Delay time.Duration
	cfg   config.Config
	repo  repositories.FileRepo
	files storage.File
	log   *slog.Logger
}

func (fw *FileWorker) Start(ctx context.Context) {
	const fn = "services.worker.FileWorker.Start"
	log := fw.log.With(slog.String("fn", fn))

	ticker := time.NewTicker(fw.Delay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("file worker stopped")
			return

		case <-ticker.C:
			aliases, err := fw.repo.DeleteExpiredFiles(ctx)
			if err != nil {
				log.Error("failed to delete expired files", sl.Error(err))
			}

			for _, alias := range aliases {
				if err := fw.files.Delete(alias); err != nil {
					log.Warn("failed to delete file from storage", sl.Error(err))
				}
			}

			if len(aliases) > 0 {
				log.Info(fmt.Sprintf("deleted %d expired files", len(aliases)))
				continue
			}

			log.Debug("deleted 0 expired files")
		}
	}
}

func NewFileWorker(repo repositories.FileRepo, files storage.File, log *slog.Logger, cfg config.Config) *FileWorker {
	return &FileWorker{
		Delay: cfg.FileWorkerDelay,
		cfg:   cfg,
		repo:  repo,
		files: files,
		log:   log,
	}
}
