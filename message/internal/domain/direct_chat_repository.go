package domain

import (
	"context"

	"github.com/google/uuid"
)

// DirectChatRepository describes direct chat persistence required by the domain and application.
type DirectChatRepository interface {
	FindByUuid(ctx context.Context, chatUuid uuid.UUID) (*DirectChat, error)
	FindAllByMemberUuid(ctx context.Context, memberUuid uuid.UUID) ([]DirectChat, error)
	Create(ctx context.Context, chat DirectChat) error
}
