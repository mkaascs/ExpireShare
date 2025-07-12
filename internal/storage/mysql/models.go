package mysql

import "time"

type File struct {
	FilePath      string
	Alias         string
	DownloadsLeft int16
	LoadedAt      time.Time
	ExpiresAt     time.Time
}

type UploadFileCommand struct {
	FilePath     string
	Alias        string
	MaxDownloads int16
	TTL          time.Duration
}
