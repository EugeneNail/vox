package domain

import (
	"context"

	"github.com/google/uuid"
)

// ChatRepository describes chat persistence required by the domain and application.
type ChatRepository interface {
	FindByUuid(ctx context.Context, chatUuid uuid.UUID) (*Chat, error)
	Create(ctx context.Context, chat Chat) error
}
