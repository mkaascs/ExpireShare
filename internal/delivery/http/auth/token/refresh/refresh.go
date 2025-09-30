package refresh

import (
	"context"
	"expire-share/internal/delivery/middlewares"
	"expire-share/internal/domain/models"
	"expire-share/internal/lib/api/response"
	"expire-share/internal/lib/log/sl"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

type Request struct {
	RefreshToken string `json:"refresh_token" required:"true"`
}

type Response struct {
	response.Response
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type TokenRefresher interface {
	RefreshToken(ctx context.Context, refreshToken string) (*models.TokenPair, error)
}

func New(refresher TokenRefresher, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http.auth.token.refresh.New"
		log = slog.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		var request Request
		request, _ = middlewares.GetParsedBodyRequest[Request](r)

		ctx := r.Context()
		tokens, err := refresher.RefreshToken(ctx, request.RefreshToken)

		if err != nil {
			if response.RenderUserServiceError(w, r, err) {
				log.Info("failed to refresh token")
				return
			}

			log.Error("failed to refresh token", sl.Error(err))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"failed to refresh token")
			return
		}

		log.Info("user login successfully")
		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		})
	}
}
