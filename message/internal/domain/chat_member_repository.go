package domain

import (
	"context"

	"github.com/google/uuid"
)

// ChatMemberRepository describes chat membership persistence required by the domain and application.
type ChatMemberRepository interface {
	FindByChatUuidAndUserUuid(ctx context.Context, chatUuid uuid.UUID, userUuid uuid.UUID) (*ChatMember, error)
	FindAllByChatUuid(ctx context.Context, chatUuid uuid.UUID) ([]ChatMember, error)
}
