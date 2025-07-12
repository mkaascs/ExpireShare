package middlewares

import (
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"time"
)

func NewLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(slog.String("component", "middleware/logger"))
		log.Info("logger middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)

			wrw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			currentTime := time.Now()
			defer func() {
				entry.Info("request completed",
					slog.Int("status_code", wrw.Status()),
					slog.Int("bytes", wrw.BytesWritten()),
					slog.String("duration", time.Since(currentTime).String()),
				)
			}()

			next.ServeHTTP(wrw, r)
		}

		return http.HandlerFunc(fn)
	}
}
