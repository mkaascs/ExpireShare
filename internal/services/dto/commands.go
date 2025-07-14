package dto

import (
	"io"
	"time"
)

type UploadFileCommand struct {
	File         io.Reader
	FileSize     int64
	Filename     string
	MaxDownloads int16
	TTL          time.Duration
}
