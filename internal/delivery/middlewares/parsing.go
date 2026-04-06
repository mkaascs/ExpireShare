package middlewares

import (
	"context"
	"expire-share/internal/config"
	"expire-share/internal/delivery/response"
	"expire-share/internal/lib/log/sl"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
)

const (
	requestField = "request"
	maxBytes     = 10 * 1024 * 1024
)

type DefaultSetter interface {
	SetDefault(cfg config.Service)
}

func NewBodyParser[T any](cfg config.Service, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		logger := log.With(slog.String("component", "middleware/parsing"))

		methodsToSkip := map[string]bool{
			http.MethodGet:     true,
			http.MethodHead:    true,
			http.MethodDelete:  true,
			http.MethodOptions: true,
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

			if methodsToSkip[r.Method] {
				next.ServeHTTP(w, r)
				return
			}

			if r.ContentLength <= 0 {
				logger.Info("request body is empty")
				response.RenderError(w, r,
					http.StatusBadRequest,
					"request body is empty")
				return
			}

			var request T
			err := render.DecodeJSON(r.Body, &request)
			if err != nil {
				logger.Info("failed to decode request body", sl.Error(err))
				response.RenderError(w, r,
					http.StatusBadRequest,
					"failed to decode request body")
				return
			}

			if defaulter, ok := any(&request).(DefaultSetter); ok {
				defaulter.SetDefault(cfg)
			}

			ctx := context.WithValue(r.Context(), requestField, request)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetParsedBodyRequest[T any](r *http.Request) (T, bool) {
	request, ok := r.Context().Value(requestField).(T)
	return request, ok
}
