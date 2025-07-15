package worker

import (
	"context"
	"expire-share/internal/config"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/services/interfaces"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type FileWorker struct {
	Delay time.Duration
	cfg   config.Config
	repo  interfaces.FileRepo
	log   *slog.Logger
}

func (fw *FileWorker) Start(ctx context.Context) {
	const fn = "services.worker.FileWorker.Start"
	fw.log = slog.With(slog.String("fn", fn))

	erg, ctx := errgroup.WithContext(ctx)

	erg.Go(func() error {
		ticker := time.NewTicker(fw.Delay)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
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
			}
		}
	})

	if err := erg.Wait(); err != nil {
		fw.log.Error("failed to run worker", sl.Error(err))
	}
}

func NewFileWorker(repo interfaces.FileRepo, log *slog.Logger, cfg config.Config) *FileWorker {
	return &FileWorker{
		Delay: cfg.FileWorkerDelay,
		cfg:   cfg,
		repo:  repo,
		log:   log,
	}
}
