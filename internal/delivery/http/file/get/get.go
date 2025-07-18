package get

import (
	"errors"
	"expire-share/internal/delivery/interfaces"
	"expire-share/internal/lib/api/response"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/services"
	"expire-share/internal/services/dto"
	"fmt"
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
	DownloadsLeft int16  `json:"downloads_left"`
	ExpiresIn     string `json:"expires_in"`
}

// New @Summary Get file info
// @Description Get info about uploaded file by its alias
// @Tags file
// @Accept json
// @Produce json
// @Param request body Request true "File data"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 403 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /file [get]
func New(fileService interfaces.FileService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http.file.get.New"
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

		command := dto.GetFileCommand{
			Alias:    alias,
			Password: request.Password,
		}

		ctx := r.Context()
		file, err := fileService.GetFileByAlias(ctx, command)
		if err != nil {
			if errors.Is(err, services.ErrAliasNotFound) {
				log.Info("file with current alias not found", slog.String("alias", alias))
				response.RenderError(w, r,
					http.StatusNotFound, "file with current alias not found")
				return
			}

			if errors.Is(err, services.ErrPasswordRequired) {
				log.Info("password is required", slog.String("alias", alias))
				response.RenderError(w, r,
					http.StatusUnauthorized, "password is required")
				return
			}

			if errors.Is(err, services.ErrIncorrectPassword) {
				log.Info("incorrect password", slog.String("alias", alias))
				response.RenderError(w, r,
					http.StatusForbidden, "incorrect password")
				return
			}

			log.Error("failed to get file info", sl.Error(err))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"failed to get file info")
			return
		}

		res := Response{
			Response:      response.Response{},
			DownloadsLeft: file.DownloadsLeft,
			ExpiresIn: fmt.Sprintf("%02d:%02d:%02d",
				int(file.ExpiresIn.Hours()), int(file.ExpiresIn.Minutes())%60, int(file.ExpiresIn.Seconds())%60),
		}

		log.Info("file info was sent", slog.String("alias", alias))
		render.Status(r, http.StatusOK)
		render.JSON(w, r, res)
	}
}
