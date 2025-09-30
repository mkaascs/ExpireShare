package login

import (
	"context"
	"expire-share/internal/delivery/middlewares"
	"expire-share/internal/domain/models"
	"expire-share/internal/lib/api/response"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/services/dto/commands"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

type Request struct {
	Login    string `json:"login" validate:"min=3" required:"true" example:"user"`
	Password string `json:"password" validate:"min=5" example:"expire123" required:"true"`
}

type Response struct {
	response.Response
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type UserLogin interface {
	Login(ctx context.Context, command commands.LoginCommand) (*models.TokenPair, error)
}

func New(login UserLogin, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http.auth.login.New"
		log = slog.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		var request Request
		request, _ = middlewares.GetParsedBodyRequest[Request](r)

		ctx := r.Context()
		tokens, err := login.Login(ctx, commands.LoginCommand{
			Login:    request.Login,
			Password: request.Password,
		})

		if err != nil {
			if response.RenderUserServiceError(w, r, err) {
				log.Info("failed to login", sl.Error(err))
				return
			}

			log.Error("failed to login", sl.Error(err))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"failed to login")
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
