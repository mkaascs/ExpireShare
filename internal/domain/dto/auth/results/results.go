package results

import "expire-share/internal/domain/entities"

type Login struct {
	Tokens    entities.TokenPair
	User      entities.User
	ExpiresIn int64
}

type Refresh struct {
	Tokens    entities.TokenPair
	ExpiresIn int64
}

type Register struct {
	UserID int64
}

type Validate struct {
	Status    entities.TokenStatus
	UserID    int64
	Roles     []entities.UserRole
	ExpiresAt int64
}
