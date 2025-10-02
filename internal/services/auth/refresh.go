package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"expire-share/internal/domain/errors/repository"
	"expire-share/internal/domain/errors/services/auth"
	"expire-share/internal/domain/models"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

func (as *Service) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenPair, error) {
	const fn = "services.auth.Service.RefreshToken"
	as.log = slog.With(slog.String("fn", fn))

	tokenHash := hmac.New(sha256.New, as.secrets.HmacSecret)
	tokenHash.Write([]byte(refreshToken))

	tokenHashStr := base64.URLEncoding.EncodeToString(tokenHash.Sum(nil))

	token, err := as.tokenRepo.GetToken(ctx, tokenHashStr)
	if err != nil {
		if errors.Is(err, repository.ErrTokenNotFound) {
			as.log.Info("token not found")
			return nil, auth.ErrTokenNotFound
		}

		as.log.Error("failed to get token")
		return nil, err
	}

	if token.ExpiresAt.Before(time.Now()) {
		as.log.Info("token expired", slog.Int64("user_id", token.UserId))
		return nil, auth.ErrTokenExpired
	}

	user, err := as.userRepo.GetUserById(ctx, token.UserId)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			as.log.Info("user not found")
			return nil, auth.ErrUserNotFound
		}

		as.log.Error("failed to get user")
		return nil, err
	}

	pair, err := as.generateTokenPair(user.Id, user.Role)
	if err != nil {
		as.log.Error("failed to generate token pair")
		return nil, err
	}

	newTokenHash, err := bcrypt.GenerateFromPassword([]byte(pair.RefreshToken), bcrypt.DefaultCost)
	if err != nil {
		as.log.Error("failed to encrypt token")
		return nil, err
	}

	err = as.tokenRepo.ReplaceToken(ctx, user.Id, string(newTokenHash))
	if err != nil {
		as.log.Error("failed to update token")
		return nil, err
	}

	return pair, nil
}
