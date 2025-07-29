package middlewares

import (
	"errors"
	"expire-share/internal/lib/api/response"
	"expire-share/internal/lib/log/sl"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
)

var validate *validator.Validate

func NewValidator(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		validate = validator.New()
		log = log.With(slog.String("component", "middleware/validating"))

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			request, ok := GetParsedBodyRequest[interface{}](r)
			if !ok {
				log.Error("failed to get parsed body request. try execute NewBodyParser before validating")
				response.RenderError(w, r,
					http.StatusInternalServerError,
					"internal server error")
				return
			}

			if err := validate.Struct(request); err != nil {
				var validationErrors validator.ValidationErrors
				errors.As(err, &validationErrors)
				log.Info("invalid request", sl.Error(err))
				response.RenderValidationError(w, r, validationErrors)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
