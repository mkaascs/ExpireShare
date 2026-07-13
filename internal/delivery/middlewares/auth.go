package middlewares

import (
	"context"
	"errors"
	"expire-share/internal/delivery/util"
	"expire-share/internal/delivery/util/response"
	"expire-share/internal/domain/dto/auth/commands"
	"expire-share/internal/domain/dto/auth/results"
	"expire-share/internal/domain/entities"
	"expire-share/internal/lib/log/sl"
	"log/slog"
	"net/http"
	"strings"
)

type authCtxKey string

const (
	userIDKey authCtxKey = "user_id"
	rolesKey  authCtxKey = "roles"
)

type UserClaims struct {
	UserID int64
	Roles  []entities.UserRole
}

type TokenValidator interface {
	ValidateToken(ctx context.Context, command commands.Validate) (*results.Validate, error)
}

func NewAuth(auth TokenValidator, log *slog.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		logger := log.With(slog.String("component", "middleware/auth"))

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r.Header.Get("Authorization"))
			if token == "" {
				logger.Info("unauthorized request")
				response.RenderError(w, r,
					http.StatusUnauthorized,
					"unauthorized request")
				return
			}

			tokenInfo, err := auth.ValidateToken(r.Context(), commands.Validate{
				AccessToken: token,
			})

			if err != nil {
				if response.RenderAuthServiceError(w, r, err) {
					logger.Info("unauthorized request", sl.Error(err))
					return
				}

				if util.IsCtxError(err) {
					return
				}

				logger.Error("failed to validate token", sl.Error(err))
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, tokenInfo.UserID)
			ctx = context.WithValue(ctx, rolesKey, tokenInfo.Roles)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func WithUserClaims(ctx context.Context, claims UserClaims) context.Context {
	ctx = context.WithValue(ctx, userIDKey, claims.UserID)
	ctx = context.WithValue(ctx, rolesKey, claims.Roles)
	return ctx
}

func GetUserClaims(r *http.Request) (*UserClaims, error) {
	userID, ok := r.Context().Value(userIDKey).(int64)
	if !ok {
		return nil, errors.New("failed to get user id from context")
	}

	roles, ok := r.Context().Value(rolesKey).([]entities.UserRole)
	if !ok {
		return nil, errors.New("failed to get user roles from context")
	}

	return &UserClaims{
		UserID: userID,
		Roles:  roles,
	}, nil
}

func extractBearerToken(header string) string {
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
