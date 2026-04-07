package domain

import (
	"context"

	"github.com/google/uuid"
)

// ProfileRepository describes profile persistence required by the domain and application.
type ProfileRepository interface {
	FindByUserUuid(ctx context.Context, userUuid uuid.UUID) (*Profile, error)
	Create(ctx context.Context, profile Profile) error
	Update(ctx context.Context, profile Profile) error
}
