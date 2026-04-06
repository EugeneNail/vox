package domain

import (
	"context"

	"github.com/google/uuid"
)

// ChatMemberRepository describes chat membership persistence required by the domain and application.
type ChatMemberRepository interface {
	FindByChatUuidAndMemberUuid(ctx context.Context, chatUuid uuid.UUID, memberUuid uuid.UUID) (*ChatMember, error)
	Create(ctx context.Context, chatMember ChatMember) error
}
