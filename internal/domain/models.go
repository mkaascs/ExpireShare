package domain

import "time"

type File struct {
	FilePath      string
	Alias         string
	DownloadsLeft int16
	PasswordHash  string
	LoadedAt      time.Time
	ExpiresAt     time.Time
	UserId        int64
}

type User struct {
	Id int64
	IP string
}
