package auth

import (
	"crypto/rand"
	"encoding/base64"
	"expire-share/internal/domain/models"
	"github.com/golang-jwt/jwt"
	"time"
)

const refreshTokenLength = 32

func (as *Service) generateTokenPair(userId int64, role models.UserRole) (*models.TokenPair, error) {
	accessJwt := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userId,
		"role": role,
		"iss":  as.cfg.Issuer,
		"exp":  time.Now().Add(as.cfg.AccessTokenTtl).Unix(),
		"iat":  time.Now().Unix(),
	})

	accessToken, err := accessJwt.SignedString(as.privateKey)
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
