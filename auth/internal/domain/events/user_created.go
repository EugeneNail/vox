package events

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserCreated is emitted after auth successfully creates a user.
type UserCreated struct {
	UserUuid  uuid.UUID `json:"userUuid"`
	CreatedAt time.Time `json:"createdAt"`
}

// UserCreatedPublisher publishes user-created events.
type UserCreatedPublisher interface {
	Publish(ctx context.Context, event UserCreated) error
}
