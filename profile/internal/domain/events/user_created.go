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

// UserCreatedConsumer starts consumption of user-created events.
type UserCreatedConsumer interface {
	ListenAndConsume(ctx context.Context)
}
