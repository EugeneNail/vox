package domain

import (
	"context"

	"github.com/google/uuid"
)

// MessageRepository describes message persistence required by the domain and application.
type MessageRepository interface {
	FindLastByChatUuid(ctx context.Context, chatUuid uuid.UUID, length int) ([]Message, error)
	Create(ctx context.Context, message Message) error
}
