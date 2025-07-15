package get

import (
	"errors"
	"expire-share/internal/delivery/interfaces"
	"expire-share/internal/lib/api/response"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/services"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

type Response struct {
	response.Response
	DownloadsLeft int16  `json:"downloads_left"`
	ExpiresIn     string `json:"expires_in"`
}

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

		ctx := r.Context()
		file, err := fileService.GetFileByAlias(ctx, alias)
		if err != nil {
			if errors.Is(err, services.ErrAliasNotFound) {
				log.Info("file with current alias not found", slog.String("alias", alias))
				response.RenderError(w, r,
					http.StatusNotFound, "file with current alias not found")
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
