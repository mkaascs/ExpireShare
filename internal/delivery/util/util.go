package util

import (
	"context"
	"errors"
)

func IsCtxError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
