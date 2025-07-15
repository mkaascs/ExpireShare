package services

import (
	"context"
	"errors"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/repository"
	"expire-share/internal/services/dto"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func (fs *FileService) DownloadFile(ctx context.Context, alias string) (*dto.DownloadFileResult, error) {
	const fn = "services.FileService.DownloadFile"
	fs.log = slog.With(slog.String("fn", fn))

	fileInfo, err := fs.repo.GetFileByAlias(ctx, alias)
	if err != nil {
		if errors.Is(err, repository.ErrAliasNotFound) {
			fs.log.Info("failed to get file info", sl.Error(err))
			return nil, ErrAliasNotFound
		}

		fs.log.Error("failed to get file info", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to get file info: %w", fn, err)
	}

	downloadsLeft, err := fs.repo.DecrementDownloadsByAlias(ctx, alias)
	if err != nil {
		if errors.Is(err, repository.ErrAliasNotFound) {
			fs.log.Info("failed decrement downloads left", sl.Error(err))
			return nil, ErrAliasNotFound
		}

		fs.log.Error("failed to decrement downloads left", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to decrement downloads left: %w", fn, err)
	}

	filePath := filepath.Join(fs.cfg.Path, fileInfo.FilePath)

	file, err := os.Open(filePath)
	if err != nil {
		fs.log.Error("failed to open file", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to open file: %w", fn, err)
	}

	stat, err := file.Stat()
	if err != nil {
		fs.log.Error("failed to stat file", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to stat file: %w", fn, err)
	}

	closeFunc := func() error {
		err := file.Close()
		if err != nil {
			fs.log.Error("failed to close file", sl.Error(err))
			return fmt.Errorf("%s: failed to close file: %w", fn, err)
		}

		return nil
	}

	if downloadsLeft > 0 {
		res := dto.DownloadFileResult{
			File:     file,
			FileInfo: stat,
			Close:    closeFunc,
		}

		return &res, nil
	}

	closeAndDeleteFunc := func() error {
		err := closeFunc()
		if err != nil {
			return err
		}

		err = os.RemoveAll(filepath.Join(fs.cfg.Path, alias))
		if err != nil {
			fs.log.Error("failed to remove file", sl.Error(err))
			return fmt.Errorf("%s: failed to remove file: %w", fn, err)
		}

		err = fs.repo.DeleteFile(ctx, alias)
		if err != nil {
			fs.log.Error("failed to delete file from repository", sl.Error(err))
			return fmt.Errorf("%s: failed to delete file from repository: %w", fn, err)
		}

		return nil
	}

	res := dto.DownloadFileResult{
		File:     file,
		FileInfo: stat,
		Close:    closeAndDeleteFunc,
	}

	return &res, nil
}
