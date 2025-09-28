package remove

import (
	"context"
	"expire-share/internal/lib/api/response"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/services/dto/commands"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

type Response struct {
	response.Response
}

type FileDeleter interface {
	DeleteFile(ctx context.Context, command commands.DeleteFileCommand) error
}

// New @Summary Delete file
// @Description Deletes uploaded file by its alias
// @Tags file
// @Accept json
// @Produce json
// @Success 204
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 403 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /file [delete]
func New(deleter FileDeleter, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http.file.api.delete.New"
		log = slog.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		alias := chi.URLParam(r, "alias")

		password := r.Header.Get("X-Resource-Password")

		ctx := r.Context()
		err := deleter.DeleteFile(ctx, commands.DeleteFileCommand{
			Alias:    alias,
			Password: password,
		})

		if err != nil {
			if response.RenderFileServiceError(w, r, err) {
				log.Info("failed to delete file info", sl.Error(err), slog.String("alias", alias))
				return
			}

			log.Error("failed to delete file", sl.Error(err), slog.String("alias", alias))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"failed to delete file")
			return
		}

		render.Status(r, http.StatusNoContent)
	}
}
