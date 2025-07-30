package repository

import "time"

type AddFileCommand struct {
	FilePath     string
	Alias        string
	MaxDownloads int16
	PasswordHash string
	TTL          time.Duration
}

type SaveTokenCommand struct {
	UserId           int64
	RefreshTokenHash string
	ExpiresAt        time.Time
}

type CheckUserCommand struct {
	Login        string
	PasswordHash string
}
