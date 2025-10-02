package auth

import (
	"crypto/rsa"
	"expire-share/internal/config"
	"expire-share/internal/services/interfaces"
	"log/slog"
)

type Secrets struct {
	PrivateKey *rsa.PrivateKey
	HmacSecret []byte
}

type Service struct {
	tokenRepo interfaces.TokenRepo
	userRepo  interfaces.UserRepo
	cfg       config.Config
	log       *slog.Logger
	secrets   Secrets
}

func New(tokenRepo interfaces.TokenRepo, userRepo interfaces.UserRepo, cfg config.Config, log *slog.Logger, secrets Secrets) *Service {
	return &Service{
		tokenRepo: tokenRepo,
		userRepo:  userRepo,
		cfg:       cfg,
		log:       log,
		secrets:   secrets,
	}
}
