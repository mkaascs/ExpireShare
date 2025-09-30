package auth

import (
	"context"
	"errors"
	"expire-share/internal/domain/errors/repository"
	"expire-share/internal/domain/errors/services/auth"
	"expire-share/internal/domain/models"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/services/dto/commands"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

func (as *Service) Login(ctx context.Context, command commands.LoginCommand) (*models.TokenPair, error) {
	const fn = "services.auth.Service.Login"
	as.log = slog.With(slog.String("fn", fn))

	user, err := as.userRepo.GetUserByLogin(ctx, command.Login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			as.log.Info("failed to get user", sl.Error(err))
			return nil, auth.ErrUserNotFound
		}

		as.log.Error("failed to get user", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to get user: %w", fn, err)
	}

	err = verifyPassword(user.PasswordHash, command.Password)
	if err != nil {
		as.log.Info("invalid password", sl.Error(err))
		return nil, auth.ErrInvalidPassword
	}

	tokens, err := as.generateTokenPair(user.Id, user.Role)
	if err != nil {
		as.log.Error("failed to generate token pair", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to generate token pair: %w", fn, err)
	}

	err = as.saveRefreshToken(ctx, user.Id, tokens.RefreshToken)
	if err != nil {
		as.log.Error("failed to save refresh token", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to save refresh token: %w", fn, err)
	}

	return tokens, nil
}

func verifyPassword(passwordHash string, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return auth.ErrInvalidPassword
	}

	return nil
}
