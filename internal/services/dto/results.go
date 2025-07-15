package dto

import (
	"os"
)

type DownloadFileResult struct {
	File     *os.File
	FileInfo os.FileInfo
	Close    func() error
}
