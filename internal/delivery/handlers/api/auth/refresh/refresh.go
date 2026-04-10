package refresh

import (
	"context"
	"expire-share/internal/delivery/middlewares"
	"expire-share/internal/delivery/util"
	"expire-share/internal/delivery/util/response"
	"expire-share/internal/domain/dto/auth/commands"
	"expire-share/internal/domain/dto/auth/results"
	"expire-share/internal/lib/log/sl"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

// Request represents token refresh request body
//
//	@Description	Refresh token for obtaining new access token
type Request struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Response represents token refresh response
//
//	@Description	New access and refresh tokens
type Response struct {
	response.Response
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`

	// Token expiration time in seconds
	//	@example	900
	ExpiresIn int64 `json:"expires_in,omitempty"`
}

type TokenRefresh interface {
	Refresh(ctx context.Context, command commands.Refresh) (*results.Refresh, error)
}

// New @Summary Refresh access token
//
//	@Description	Get new access token using refresh token. Performs token rotation (returns new refresh token).
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		Request				true	"Refresh token"
//	@Success		200		{object}	Response			"Token refreshed successfully"
//	@Failure		400		{object}	response.Response	"Invalid request body"
//	@Failure		401		{object}	response.Response	"Invalid or expired refresh token"
//	@Failure		422		{object}	response.Response	"Validation error"
//	@Failure		500		{object}	response.Response	"Internal server error"
//	@Router			/api/auth/refresh [post]
func New(refresh TokenRefresh, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http.api.auth.refresh.New"
		log := log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		request, ok := middlewares.GetParsedBodyRequest[Request](r)
		if !ok {
			log.Error("failed to parse request")
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"internal server error")
			return
		}

		result, err := refresh.Refresh(r.Context(), commands.Refresh{
			RefreshToken: request.RefreshToken,
		})

		if err != nil {
			if response.RenderAuthServiceError(w, r, err) || util.IsCtxError(err) {
				log.Info("failed to refresh token", sl.Error(err))
				return
			}

			log.Error("failed to refresh token", sl.Error(err))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"internal server error")
			return
		}

		log.Info("user refresh token successfully")
		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			AccessToken:  result.Tokens.AccessToken,
			RefreshToken: result.Tokens.RefreshToken,
			ExpiresIn:    result.ExpiresIn,
		})
	}
}
