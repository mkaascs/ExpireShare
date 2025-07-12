package mysql

import "time"

type UploadFileCommand struct {
	FilePath     string
	Alias        string
	MaxDownloads int16
	TTL          time.Duration
}
