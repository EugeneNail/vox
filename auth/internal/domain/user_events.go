package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserCreatedEvent is emitted after auth successfully creates a user.
type UserCreatedEvent struct {
	UserUuid  uuid.UUID `json:"userUuid"`
	CreatedAt time.Time `json:"createdAt"`
}

// UserCreatedPublisher publishes user-created events.
type UserCreatedPublisher interface {
	Publish(ctx context.Context, event UserCreatedEvent) error
}
