package login

import (
	"context"
	"expire-share/internal/delivery/middlewares"
	"expire-share/internal/delivery/response"
	"expire-share/internal/delivery/util"
	"expire-share/internal/domain/dto/auth/commands"
	"expire-share/internal/domain/dto/auth/results"
	"expire-share/internal/lib/log/sl"
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

type UserLogin interface {
	Login(ctx context.Context, command commands.Login) (*results.Login, error)
}

func New(login UserLogin, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http.auth.login.New"
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

		result, err := login.Login(r.Context(), commands.Login{
			Login:    request.Login,
			Password: request.Password,
		})

		if err != nil {
			if response.RenderUserServiceError(w, r, err) || util.IsCtxError(err) {
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
