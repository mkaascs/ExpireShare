package files

import "errors"

var (
	ErrFileSizeTooBig    = errors.New("file size too big")
	ErrAliasNotFound     = errors.New("file with current alias does not exist")
	ErrPasswordRequired  = errors.New("password required for access")
	ErrIncorrectPassword = errors.New("incorrect password")
)
