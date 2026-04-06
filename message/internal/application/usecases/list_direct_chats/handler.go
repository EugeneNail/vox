package list_direct_chats

import (
	"context"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
)

// Handler lists direct chats through the list_direct_chats use-case.
type Handler struct {
	chatRepository domain.ChatRepository
}

// Query contains the input required to list direct chats.
type Query struct {
	UserUuid uuid.UUID
}

// NewHandler constructs a list_direct_chats handler with its dependencies.
func NewHandler(chatRepository domain.ChatRepository) *Handler {
	return &Handler{
		chatRepository: chatRepository,
	}
}

// Handle returns direct chats available to the user.
func (handler *Handler) Handle(ctx context.Context, query Query) ([]domain.Chat, error) {
	chats, err := handler.chatRepository.FindAllDirectByMemberUuid(ctx, query.UserUuid)
	if err != nil {
		return nil, fmt.Errorf("finding direct chats by user uuid %q: %w", query.UserUuid, err)
	}

	return chats, nil
}
