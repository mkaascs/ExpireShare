package remove

import (
	"errors"
	"expire-share/internal/delivery/interfaces"
	"expire-share/internal/lib/api/response"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/services/dto"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

type Request struct {
	Password string `json:"password,omitempty" example:"1234"`
}

type Response struct {
	response.Response
}

// New @Summary Delete file
// @Description Deletes uploaded file by its alias
// @Tags file
// @Accept json
// @Produce json
// @Param request body Request true "File data"
// @Success 204
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 403 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /file [delete]
func New(fileService interfaces.FileService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http.file.delete.New"
		log = slog.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("no alias provided")
			response.RenderError(w, r,
				http.StatusBadRequest,
				"no alias provided")
			return
		}

		var request Request
		err := render.DecodeJSON(r.Body, &request)
		if err != nil {
			log.Info("failed to decode json body", sl.Error(err))
			request.Password = ""
		}

		command := dto.DeleteFileCommand{
			Alias:    alias,
			Password: request.Password,
		}

		ctx := r.Context()
		err = fileService.DeleteFile(ctx, command)
		if err != nil {
			if errors.Is(err, files.ErrAliasNotFound) {
				log.Info("file with current alias not found", slog.String("alias", alias))
				response.RenderError(w, r,
					http.StatusNotFound,
					"file with current alias not found")
				return
			}

			if errors.Is(err, files.ErrPasswordRequired) {
				log.Info("password is required", slog.String("alias", alias))
				response.RenderError(w, r,
					http.StatusUnauthorized, "password is required")
				return
			}

			if errors.Is(err, files.ErrIncorrectPassword) {
				log.Info("incorrect password", slog.String("alias", alias))
				response.RenderError(w, r,
					http.StatusForbidden, "incorrect password")
				return
			}

			log.Error("failed to delete file", sl.Error(err))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"failed to delete file")
			return
		}

		render.Status(r, http.StatusNoContent)
	}
}
