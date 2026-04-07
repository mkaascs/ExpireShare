package repositories

import (
	"context"
)

type TxBeginner interface {
	BeginTx(ctx context.Context) (Tx, error)
}

type Tx interface {
	Commit() error
	Rollback() error
}
