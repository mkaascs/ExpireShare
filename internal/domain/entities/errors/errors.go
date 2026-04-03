package errors

import "errors"

var (
	ErrAliasExists     = errors.New("current alias already exists")
	ErrAliasNotFound   = errors.New("file with current alias does not exist")
	ErrNoDownloadsLeft = errors.New("there is no downloads left")
	ErrFileSizeTooBig  = errors.New("file size too big")

	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidCredentials = errors.New("invalid login or password")

	ErrFilePasswordRequired = errors.New("file password required for access")
	ErrFilePasswordInvalid  = errors.New("invalid file password")
)
