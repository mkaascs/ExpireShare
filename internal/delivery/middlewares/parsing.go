package middlewares

import (
	"context"
	"expire-share/internal/lib/api/response"
	"expire-share/internal/lib/log/sl"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

const (
	fieldName = "request"
	maxBytes  = 10 * 1024 * 1024
)

func NewBodyParser(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(slog.String("component", "middleware/parsing"))

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
				log.Info("request body is empty")
				response.RenderError(w, r,
					http.StatusBadRequest,
					"request body is empty")
				return
			}

			var request interface{}
			err := render.DecodeJSON(r.Body, &request)
			if err != nil {
				log.Info("failed to decode request body", sl.Error(err))
				response.RenderError(w, r,
					http.StatusBadRequest,
					"failed to decode request body")
				return
			}

			ctx := context.WithValue(r.Context(), fieldName, request)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetParsedBodyRequest[T any](r *http.Request) (T, bool) {
	request, ok := r.Context().Value(fieldName).(T)
	return request, ok
}
