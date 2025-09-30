package auth

import (
	"context"
	"errors"
	"expire-share/internal/domain/errors/repository"
	"expire-share/internal/domain/errors/services/auth"
	"expire-share/internal/domain/models"
	"expire-share/internal/lib/log/sl"
	"expire-share/internal/services/dto/commands"
	dto "expire-share/internal/services/dto/repository"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

func (as *Service) Register(ctx context.Context, command commands.RegisterCommand) (*models.TokenPair, error) {
	const fn = "services.auth.Service.Register"
	as.log = slog.With(slog.String("fn", fn))

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(command.Password), bcrypt.DefaultCost)
	if err != nil {
		as.log.Error("failed to hash password", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to hash password: %w", fn, err)
	}

	userId, err := as.userRepo.AddUser(ctx, dto.AddUserCommand{
		Login:        command.Login,
		PasswordHash: string(hashedPassword),
	})

	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			as.log.Info("failed to register user", sl.Error(err))
			return nil, auth.ErrUserAlreadyExists
		}

		as.log.Error("failed to register user", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to register user: %w", fn, err)
	}

	tokens, err := as.generateTokenPair(userId, models.Client)
	if err != nil {
		as.log.Error("failed to generate token pair", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to generate token pair: %w", fn, err)
	}

	err = as.saveRefreshToken(ctx, userId, tokens.RefreshToken)
	if err != nil {
		as.log.Error("failed to save refresh token", sl.Error(err))
		return nil, fmt.Errorf("%s: failed to save refresh token: %w", fn, err)
	}

	return tokens, nil
}
