package middlewares

import (
	"expire-share/internal/lib/api/response"
	"expire-share/internal/services/auth"
	"log/slog"
	"net/http"
)

func NewAuth(authService *auth.Service, log *slog.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		log = log.With(slog.String("component", "middleware/auth"))

		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Info("authorization header is empty")
				w.Header().Set("WWW-Authenticate", `Basic realm="api"`)
				response.RenderError(w, r,
					http.StatusUnauthorized,
					"unauthorized")
				return
			}
		}
	}
}
