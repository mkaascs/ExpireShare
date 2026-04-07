package repositories

import (
	"context"
	"expire-share/internal/domain/dto/files/commands"
	"expire-share/internal/domain/entities"
)

type FileRepo interface {
	TxBeginner

	GetFileByAlias(ctx context.Context, alias string) (entities.File, error)
	CountByUserID(ctx context.Context, userID int64) (int, error)

	AddFileTx(ctx context.Context, tx Tx, command commands.AddFile) (int64, error)
	DecrementDownloadsByAliasTx(ctx context.Context, tx Tx, alias string) (int16, error)
	DeleteFileTx(ctx context.Context, tx Tx, alias string) error
	DeleteExpiredFilesTx(ctx context.Context, tx Tx, limit int) ([]string, error)
}
