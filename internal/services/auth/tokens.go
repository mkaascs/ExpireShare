package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"expire-share/internal/domain/models"
	dto "expire-share/internal/services/dto/repository"
	"strconv"
	"time"
)

const refreshTokenLength = 32

func (as *Service) generateTokenPair(userId int64, role models.UserRole) (*models.TokenPair, error) {
	_, accessToken, err := as.JwtAuth.Encode(map[string]interface{}{
		"sub":  strconv.FormatInt(userId, 10),
		"role": role,
		"exp":  time.Now().Add(as.cfg.AccessTokenTtl).Unix(),
	})

	if err != nil {
		return nil, err
	}

	refreshToken, err := as.generateRefreshToken()
	if err != nil {
		return nil, err
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (as *Service) generateRefreshToken() (string, error) {
	bytes := make([]byte, refreshTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (as *Service) saveRefreshToken(ctx context.Context, userId int64, refreshToken string) error {
	refreshTokenHash := hmac.New(sha256.New, as.secrets.RefreshTokenSecret)
	refreshTokenHash.Write([]byte(refreshToken))

	_, err := as.tokenRepo.SaveToken(ctx, dto.SaveTokenCommand{
		RefreshTokenHash: base64.URLEncoding.EncodeToString(refreshTokenHash.Sum(nil)),
		ExpiresAt:        time.Now().Add(as.cfg.RefreshTokenTtl),
		UserId:           userId,
	})

	return err
}
