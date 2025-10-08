package repository

import "time"

type AddFileCommand struct {
	FilePath     string
	Alias        string
	MaxDownloads int16
	PasswordHash string
	TTL          time.Duration
	UserId       int64
}

type SaveTokenCommand struct {
	UserId           int64
	RefreshTokenHash string
	ExpiresAt        time.Time
}

type AddUserCommand struct {
	Login        string
	PasswordHash string
}
