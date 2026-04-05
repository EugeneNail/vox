package domain

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository describes user persistence required by the domain and application.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUuid(ctx context.Context, userUuid uuid.UUID) (*User, error)
	Create(ctx context.Context, user User) error
}
