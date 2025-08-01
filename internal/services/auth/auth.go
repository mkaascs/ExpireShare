package auth

import (
	"crypto/rsa"
	"expire-share/internal/config"
	"expire-share/internal/services/interfaces"
	"log/slog"
)

type Service struct {
	tokenRepo  interfaces.TokenRepo
	userRepo   interfaces.UserRepo
	cfg        config.Config
	log        *slog.Logger
	privateKey *rsa.PrivateKey
}

func New(tokenRepo interfaces.TokenRepo, userRepo interfaces.UserRepo, cfg config.Config, log *slog.Logger, privateKey *rsa.PrivateKey) *Service {
	return &Service{
		tokenRepo:  tokenRepo,
		userRepo:   userRepo,
		cfg:        cfg,
		log:        log,
		privateKey: privateKey,
	}
}
