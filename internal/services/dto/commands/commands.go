package commands

import (
	"io"
	"time"
)

type UploadFileCommand struct {
	File         io.Reader
	FileSize     int64
	Filename     string
	MaxDownloads int16
	Password     string
	TTL          time.Duration
	UserId       int64
}

type DownloadFileCommand struct {
	Alias    string
	Password string
}

type GetFileCommand struct {
	Alias    string
	Password string
}

type DeleteFileCommand struct {
	Alias    string
	Password string
}

type LoginCommand struct {
	Login    string
	Password string
}

type RegisterCommand struct {
	Login    string
	Password string
}
