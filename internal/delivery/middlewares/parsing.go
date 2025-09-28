package middlewares

import (
	"context"
	"expire-share/internal/lib/api/response"
	"expire-share/internal/lib/log/sl"
	"github.com/go-chi/render"
	"io"
	"log/slog"
	"net/http"
)

const fieldName = "request"

type BodyParserSettings struct {
	BodyIsOptional bool
}

func NewBodyParser(log *slog.Logger, settings BodyParserSettings) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(slog.String("component", "middleware/parsing"))

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var request interface{}
			err := render.DecodeJSON(r.Body, &request)
			if err != nil && (err == io.EOF && !settings.BodyIsOptional) {
				log.Info("failed to decode json body", sl.Error(err))
				response.RenderError(w, r,
					http.StatusBadRequest,
					"failed to decode json body")
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
