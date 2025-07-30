package results

import (
	"io"
	"os"
	"time"
)

type DownloadFileResult struct {
	File     io.Reader
	FileInfo os.FileInfo
	Close    func() error
}

type GetFileResult struct {
	DownloadsLeft int16
	ExpiresIn     time.Duration
}
