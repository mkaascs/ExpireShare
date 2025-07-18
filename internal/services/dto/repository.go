package dto

import "time"

type AddFileCommand struct {
	FilePath     string
	Alias        string
	MaxDownloads int16
	PasswordHash string
	TTL          time.Duration
}
