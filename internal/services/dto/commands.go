package dto

import (
	"os"
	"time"
)

type UploadFileCommand struct {
	File         *os.File
	Filename     string
	MaxDownloads int16
	TTL          time.Duration
}
