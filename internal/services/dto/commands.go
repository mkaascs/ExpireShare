package dto

import (
	"mime/multipart"
	"time"
)

type UploadFileCommand struct {
	File         multipart.File
	FileSize     int64
	Filename     string
	MaxDownloads int16
	TTL          time.Duration
}
