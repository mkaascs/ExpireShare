package grpc

import (
	"context"
	"expire-share/internal/domain/dto/auth/commands"
	"expire-share/internal/domain/dto/auth/results"
	"expire-share/internal/domain/entities"
	domainErrors "expire-share/internal/domain/entities/errors"
	authv1 "github.com/mkaascs/AuthProto/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthClient struct {
	authClient  authv1.AuthClient
	tokenClient authv1.TokenClient
}

func NewAuthClient(grpcConn *grpc.ClientConn) *AuthClient {
	return &AuthClient{
		authClient:  authv1.NewAuthClient(grpcConn),
		tokenClient: authv1.NewTokenClient(grpcConn),
	}
}

func (ac *AuthClient) Login(ctx context.Context, command commands.Login) (*results.Login, error) {
	result, err := ac.authClient.Login(ctx, &authv1.LoginRequest{
		Login:    command.Login,
		Password: command.Password,
	})

	if err != nil {
		return nil, mapGrpcError(err)
	}

	return &results.Login{
		Tokens: entities.TokenPair{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
		},

		ExpiresIn: result.ExpiresIn,
		User:      pbUserToDomain(result.User),
	}, nil
}

func (ac *AuthClient) Register(ctx context.Context, command commands.Register) (*results.Register, error) {
	result, err := ac.authClient.Register(ctx, &authv1.RegisterRequest{
		Login:    command.Login,
		Email:    command.Email,
		Password: command.Password,
	})

	if err != nil {
		return nil, mapGrpcError(err)
	}

	return &results.Register{
		UserID: result.UserId,
	}, nil
}

func (ac *AuthClient) Refresh(ctx context.Context, command commands.Refresh) (*results.Refresh, error) {
	result, err := ac.authClient.Refresh(ctx, &authv1.RefreshRequest{
		RefreshToken: command.RefreshToken,
	})

	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return nil, domainErrors.ErrInvalidRefreshToken
		}

		return nil, mapGrpcError(err)
	}

	return &results.Refresh{
		Tokens: entities.TokenPair{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
		},

		ExpiresIn: result.ExpiresIn,
	}, nil
}

func (ac *AuthClient) Logout(ctx context.Context, command commands.Logout) error {
	_, err := ac.authClient.Logout(ctx, &authv1.LogoutRequest{
		RefreshToken: command.RefreshToken,
		AccessToken:  command.AccessToken,
	})

	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return domainErrors.ErrInvalidRefreshToken
		}

		return mapGrpcError(err)
	}

	return nil
}

func (ac *AuthClient) ValidateToken(ctx context.Context, command commands.Validate) (*results.Validate, error) {
	result, err := ac.tokenClient.ValidateToken(ctx, &authv1.ValidateTokenRequest{
		AccessToken: command.AccessToken,
	})

	if err != nil {
		return nil, mapGrpcError(err)
	}

	switch result.Status {
	case authv1.TokenStatus_EXPIRED:
		return nil, domainErrors.ErrAccessTokenExpired
	case authv1.TokenStatus_INVALID:
		return nil, domainErrors.ErrInvalidAccessToken
	case authv1.TokenStatus_REVOKED:
		return nil, domainErrors.ErrAccessTokenRevoked
	}

	return &results.Validate{
		UserID:    result.UserId,
		Roles:     pbRolesToDomain(result.Roles),
		ExpiresAt: result.ExpiresAt,
	}, nil
}
