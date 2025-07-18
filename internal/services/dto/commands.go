package dto

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
}

type byAliasCommand struct {
	Alias    string
	Password string
}

type DownloadFileCommand struct {
	byAliasCommand
}

type GetFileCommand struct {
	byAliasCommand
}

type DeleteFileCommand struct {
	byAliasCommand
}
