package entities

import "time"

type File struct {
	FilePath      string
	Alias         string
	DownloadsLeft int16
	PasswordHash  string
	LoadedAt      time.Time
	ExpiresAt     time.Time
	UserID        int64
}
