package register

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
	Login    string `json:"login" validate:"required,min=3" example:"user"`
	Password string `json:"password" validate:"required,min=5" example:"expire123"`
}

type Response struct {
	response.Response
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type UserRegister interface {
	Register(ctx context.Context, command commands.RegisterCommand) (*models.TokenPair, error)
}

func New(register UserRegister, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http.auth.register.New"
		log = slog.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		var request Request
		request, _ = middlewares.GetParsedBodyRequest[Request](r)

		ctx := r.Context()
		tokens, err := register.Register(ctx, commands.RegisterCommand{
			Login:    request.Login,
			Password: request.Password,
		})

		if err != nil {
			if response.RenderUserServiceError(w, r, err) {
				log.Info("failed to register user")
				return
			}

			log.Error("failed to register user", sl.Error(err))
			response.RenderError(w, r,
				http.StatusInternalServerError,
				"failed to register")
			return
		}

		log.Info("user register successfully")
		render.Status(r, http.StatusOK)
		render.JSON(w, r, Response{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		})
	}
}
