package repository

import "errors"

var (
	ErrAliasExists     = errors.New("current alias already exists")
	ErrAliasNotFound   = errors.New("file with current alias does not exist")
	ErrNoDownloadsLeft = errors.New("there is no downloads left")
	ErrTokenExists     = errors.New("token with this user id already exists")
	ErrTokenNotFound   = errors.New("token does not exist")
)
