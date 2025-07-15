package dto

import (
	"io"
	"os"
)

type DownloadFileResult struct {
	File     io.Reader
	FileInfo os.FileInfo
	Close    func() error
}
