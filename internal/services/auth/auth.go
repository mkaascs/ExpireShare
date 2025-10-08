package auth

import (
	"expire-share/internal/config"
	"expire-share/internal/services/interfaces"
	"github.com/go-chi/jwtauth"
	"log/slog"
)

type Secrets struct {
	AccessTokenSecret  []byte
	RefreshTokenSecret []byte
}

type Service struct {
	tokenRepo interfaces.TokenRepo
	userRepo  interfaces.UserRepo
	cfg       config.Config
	log       *slog.Logger
	secrets   Secrets
	jwtAuth   *jwtauth.JWTAuth
}

func New(tokenRepo interfaces.TokenRepo, userRepo interfaces.UserRepo, cfg config.Config, log *slog.Logger, secrets Secrets) *Service {
	return &Service{
		tokenRepo: tokenRepo,
		userRepo:  userRepo,
		cfg:       cfg,
		log:       log,
		secrets:   secrets,
		jwtAuth:   jwtauth.New("HS256", []byte(secrets.AccessTokenSecret), nil),
	}
}
