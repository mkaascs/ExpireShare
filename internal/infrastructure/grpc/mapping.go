package grpc

import (
	"context"
	"expire-share/internal/domain/entities"
	domainErrors "expire-share/internal/domain/entities/errors"
	authv1 "github.com/mkaascs/AuthProto/gen/go/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func pbRolesToDomain(roles []string) []entities.UserRole {
	result := make([]entities.UserRole, 0, len(roles))
	for _, role := range roles {
		result = append(result, entities.UserRole(role))
	}

	return result
}

func pbUserToDomain(user *authv1.UserInfo) entities.User {
	return entities.User{
		ID:        user.UserId,
		Login:     user.Login,
		Roles:     pbRolesToDomain(user.Roles),
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt.AsTime(),
	}
}

func mapGrpcError(grpcErr error) error {
	if status.Code(grpcErr) == codes.Unauthenticated {
		return domainErrors.ErrInvalidCredentials
	}

	if status.Code(grpcErr) == codes.AlreadyExists {
		return domainErrors.ErrUserAlreadyExists
	}

	if status.Code(grpcErr) == codes.Canceled {
		return context.Canceled
	}

	if status.Code(grpcErr) == codes.DeadlineExceeded {
		return context.DeadlineExceeded
	}

	return grpcErr
}
