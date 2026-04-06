package list_direct_chats

import (
	"context"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
)

// Handler lists direct chats through the list_direct_chats use-case.
type Handler struct {
	directChatRepository domain.DirectChatRepository
}

// Query contains the input required to list direct chats.
type Query struct {
	UserUuid uuid.UUID
}

// NewHandler constructs a list_direct_chats handler with its dependencies.
func NewHandler(directChatRepository domain.DirectChatRepository) *Handler {
	return &Handler{
		directChatRepository: directChatRepository,
	}
}

// Handle returns direct chats available to the user.
func (handler *Handler) Handle(ctx context.Context, query Query) ([]domain.DirectChat, error) {
	chats, err := handler.directChatRepository.FindAllByMemberUuid(ctx, query.UserUuid)
	if err != nil {
		return nil, fmt.Errorf("finding direct chats by user uuid %q: %w", query.UserUuid, err)
	}

	return chats, nil
}
