package tx

import (
	"context"
)

type Beginner interface {
	BeginTx(ctx context.Context) (Tx, error)
}

type Tx interface {
	Commit() error
	Rollback() error
}
