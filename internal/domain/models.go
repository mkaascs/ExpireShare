package domain

import "time"

type File struct {
	FilePath      string
	Alias         string
	DownloadsLeft int16
	LoadedAt      time.Time
	ExpiresAt     time.Time
}
