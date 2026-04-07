package domain

import (
	"context"

	"github.com/google/uuid"
)

// MessageRepository describes message persistence required by the domain and application.
type MessageRepository interface {
	FindByUuid(ctx context.Context, messageUuid uuid.UUID) (*Message, error)
	FindLastByChatUuid(ctx context.Context, chatUuid uuid.UUID, length int) ([]Message, error)
	Create(ctx context.Context, message Message) error
	Update(ctx context.Context, message Message) error
	Delete(ctx context.Context, messageUuid uuid.UUID) error
}
