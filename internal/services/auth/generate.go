package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"expire-share/internal/domain/models"
	dto "expire-share/internal/services/dto/repository"
	"github.com/golang-jwt/jwt"
	"time"
)

const refreshTokenLength = 32

func (as *Service) saveRefreshToken(ctx context.Context, userId int64, refreshToken string) error {
	refreshTokenHash := hmac.New(sha256.New, as.secrets.HmacSecret)
	refreshTokenHash.Write([]byte(refreshToken))

	_, err := as.tokenRepo.SaveToken(ctx, dto.SaveTokenCommand{
		RefreshTokenHash: base64.URLEncoding.EncodeToString(refreshTokenHash.Sum(nil)),
		ExpiresAt:        time.Now().Add(as.cfg.RefreshTokenTtl),
		UserId:           userId,
	})

	return err
}

func (as *Service) generateTokenPair(userId int64, role models.UserRole) (*models.TokenPair, error) {
	accessJwt := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":  userId,
		"role": role,
		"iss":  as.cfg.Issuer,
		"exp":  time.Now().Add(as.cfg.AccessTokenTtl).Unix(),
		"iat":  time.Now().Unix(),
	})

	accessToken, err := accessJwt.SignedString(as.secrets.PrivateKey)
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
