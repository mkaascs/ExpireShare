package worker

import (
	"context"
	"errors"
	"expire-share/internal/config"
	"expire-share/internal/domain/interfaces/repositories"
	"expire-share/internal/domain/interfaces/storage"
	"expire-share/internal/lib/log/sl"
	"log/slog"
	"time"
)

type FileWorker struct {
	Delay time.Duration
	repo  repositories.FileRepo
	files storage.File
	log   *slog.Logger
}

const batchLimit = 100

func (fw *FileWorker) Start(ctx context.Context) {
	const fn = "services.worker.FileWorker.Start"
	log := fw.log.With(slog.String("fn", fn))

	ticker := time.NewTicker(fw.Delay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("file worker stopping")
			return

		case <-ticker.C:
			tx, err := fw.repo.BeginTx(ctx)
			if err != nil {
				log.Warn("failed to begin tx. trying again in 5s", sl.Error(err))
				time.Sleep(5 * time.Second)
				continue
			}

			rollback := func() {
				if err := tx.Rollback(); err != nil {
					log.Warn("failed to rollback tx", sl.Error(err))
				}
			}

			aliases, err := fw.repo.DeleteExpiredFilesTx(ctx, tx, batchLimit)
			if err != nil {
				log.Warn("failed to delete expired files from repo", sl.Error(err))
				rollback()
				continue
			}

			success := true
			for _, alias := range aliases {
				if err := fw.files.Delete(ctx, alias); err != nil {
					if errors.Is(err, context.Canceled) {
						log.Info("file worker stopping", slog.String("alias", alias))
						rollback()
						return
					}

					log.Warn("failed to delete file from storage", sl.Error(err), slog.String("alias", alias))
					success = false
					break
				}
			}

			if !success {
				rollback()
				continue
			}

			if err := tx.Commit(); err != nil {
				log.Warn("failed to commit tx", sl.Error(err))
				continue
			}

			if len(aliases) > 0 {
				log.Info("deleted expired files", slog.Int("count", len(aliases)))
				continue
			}

			log.Debug("deleted 0 expired files")
		}
	}
}

func NewFileWorker(repo repositories.FileRepo, files storage.File, log *slog.Logger, cfg config.Config) *FileWorker {
	return &FileWorker{
		Delay: cfg.FileWorkerDelay,
		repo:  repo,
		files: files,
		log:   log,
	}
}
