package get

import (
	"context"
	"expire-share/internal/delivery/middlewares"
	"expire-share/internal/delivery/util"
	"expire-share/internal/delivery/util/response"
	"expire-share/internal/domain/dto/files/commands"
	"expire-share/internal/domain/dto/files/results"
	"expire-share/internal/lib/log/sl"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

type Response struct {
	response.Response
	DownloadsLeft int16  `json:"downloads_left,omitempty"`
	ExpiresIn     string `json:"expires_in,omitempty"`
}

type FileGetter interface {
	GetFileByAlias(ctx context.Context, command commands.GetFile) (*results.GetFile, error)
}

// New @Summary Get file info
//
//	@Description	Get info about uploaded file by its alias
//	@Tags			file
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	Response
//	@Failure		400	{object}	Response
//	@Failure		401	{object}	Response
//	@Failure		403	{object}	Response
//	@Failure		404	{object}	Response
//	@Failure		500	{object}	Response
//	@Router			/file [get]
func New(getter FileGetter, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http.file.api.get.New"
		log := log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		alias := chi.URLParam(r, "alias")
		password := r.Header.Get("X-Resource-Password")

		claims, err := middlewares.GetUserClaims(r)
		if err != nil {
			log.Error("failed to get user claims", sl.Error(err))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"internal server error")
			return
		}

		file, err := getter.GetFileByAlias(r.Context(), commands.GetFile{
			Alias:    alias,
			Password: password,
			RequestingUserInfo: commands.RequestingUserInfo{
				UserID: claims.UserID,
				Roles:  claims.Roles,
			},
		})

		if err != nil {
			const msg = "failed to get file info by alias"
			if response.RenderFileServiceError(w, r, err) || util.IsCtxError(err) {
				log.Info(msg, sl.Error(err), slog.String("alias", alias))
				return
			}

			log.Error(msg, sl.Error(err), slog.String("alias", alias))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"internal server error")
			return
		}

		log.Info("file info was sent", slog.String("alias", alias))
		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			DownloadsLeft: file.DownloadsLeft,
			ExpiresIn: fmt.Sprintf("%02d:%02d:%02d",
				int(file.ExpiresIn.Hours()), int(file.ExpiresIn.Minutes())%60, int(file.ExpiresIn.Seconds())%60),
		})
	}
}
