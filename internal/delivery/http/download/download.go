package download

import (
	"context"
	"errors"
	"expire-share/internal/lib/api/response"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/services/dto"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"syscall"
)

type Request struct {
	Password string `json:"password,omitempty" example:"1234"`
}

type Response struct {
	response.Response
}

type FileDownloader interface {
	DownloadFile(ctx context.Context, command dto.DownloadFileCommand) (*dto.DownloadFileResult, error)
}

// New @Summary Download file
// @Description Downloads uploaded file by its alias
// @Tags file
// @Accept json
// @Produce json
// @Param request body Request true "File data"
// @Success 200
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 403 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /download [get]
func New(downloader FileDownloader, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http.download.New"
		log = slog.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		alias := chi.URLParam(r, "alias")

		var request Request
		err := render.DecodeJSON(r.Body, &request)
		if err != nil {
			log.Info("failed to decode json body", sl.Error(err))
			request.Password = ""
		}

		command := dto.DownloadFileCommand{
			Alias:    alias,
			Password: request.Password,
		}

		ctx := r.Context()
		file, err := downloader.DownloadFile(ctx, command)
		if err != nil {
			if response.RenderFileServiceError(w, r, err) {
				log.Info("failed to get file info", sl.Error(err), slog.String("alias", alias))
				return
			}

			log.Error("failed to get file info", sl.Error(err), slog.String("alias", alias))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"failed to get file")
			return
		}

		contentType := mime.TypeByExtension(filepath.Ext(file.FileInfo.Name()))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.FileInfo.Name()))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", file.FileInfo.Size()))

		data, err := io.ReadAll(file.File)
		_, err = w.Write(data)
		if err != nil {
			if err := file.Close(); err != nil {
				log.Error("failed to close file", sl.Error(err))
			}

			if errors.Is(err, syscall.EPIPE) || errors.Is(err, io.ErrClosedPipe) {
				log.Info("client disconnected during download")
				return
			}

			log.Error("failed to write response", sl.Error(err))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"failed to write response")
			return
		}

		if err := file.Close(); err != nil {
			log.Error("failed to close file", sl.Error(err))
			return
		}

		log.Info("file was successfully downloaded", slog.String("alias", alias))
	}
}
