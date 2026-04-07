package delete

import (
	"context"
	"expire-share/internal/delivery/middlewares"
	"expire-share/internal/delivery/util"
	"expire-share/internal/delivery/util/response"
	"expire-share/internal/domain/dto/files/commands"
	"expire-share/internal/lib/log/sl"
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
	DeleteFile(ctx context.Context, command commands.DeleteFile) error
}

// New @Summary Delete file
//
//	@Description	Deletes uploaded file by its alias
//	@Tags			file
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Failure		400	{object}	Response
//	@Failure		401	{object}	Response
//	@Failure		403	{object}	Response
//	@Failure		404	{object}	Response
//	@Failure		500	{object}	Response
//	@Router			/file [delete]
func New(deleter FileDeleter, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http.file.api.delete.New"
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

		err = deleter.DeleteFile(r.Context(), commands.DeleteFile{
			Alias:    alias,
			Password: password,
			RequestingUserInfo: commands.RequestingUserInfo{
				UserID: claims.UserID,
				Roles:  claims.Roles,
			},
		})

		if err != nil {
			if response.RenderFileServiceError(w, r, err) || util.IsCtxError(err) {
				log.Info("failed to delete file", sl.Error(err), slog.String("alias", alias))
				return
			}

			log.Error("failed to delete file", sl.Error(err), slog.String("alias", alias))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"internal server error")
			return
		}

		render.Status(r, http.StatusNoContent)
	}
}
