package commands

import (
	"io"
	"time"
)

type UploadFile struct {
	File         io.Reader
	FileSize     int64
	Filename     string
	MaxDownloads int16
	Password     string
	TTL          time.Duration
	UserID       int64
}

type DownloadFile struct {
	Alias    string
	Password string
}

type GetFile struct {
	Alias    string
	Password string
}

type DeleteFile struct {
	Alias    string
	Password string
}

type AddFile struct {
	FilePath     string
	Alias        string
	MaxDownloads int16
	PasswordHash string
	TTL          time.Duration
	UserID       int64
}
