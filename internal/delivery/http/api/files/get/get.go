package get

import (
	"context"
	"expire-share/internal/lib/api/response"
	"expire-share/internal/lib/log/sl"
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

type FileGetter interface {
	GetFileByAlias(ctx context.Context, command dto.GetFileCommand) (*dto.GetFileResult, error)
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
func New(getter FileGetter, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http.file.api.get.New"
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

		ctx := r.Context()
		file, err := getter.GetFileByAlias(ctx, dto.GetFileCommand{
			Alias:    alias,
			Password: request.Password,
		})

		if err != nil {
			if response.RenderFileServiceError(w, r, err) {
				log.Info("failed to get file info", sl.Error(err), slog.String("alias", alias))
				return
			}

			log.Error("failed to get file info", sl.Error(err), slog.String("alias", alias))
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
