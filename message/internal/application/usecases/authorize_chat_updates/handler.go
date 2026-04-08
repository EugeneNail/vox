package authorize_chat_updates

import (
	"context"
	"errors"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
)

var ErrChatNotFound = errors.New("chat not found")
var ErrChatAccessDenied = errors.New("chat access denied")

// Handler authorizes runtime chat update subscriptions.
type Handler struct {
	chatRepository       domain.ChatRepository
	chatMemberRepository domain.ChatMemberRepository
}

// Query contains the input required to authorize chat realtime updates.
type Query struct {
	ChatUuid uuid.UUID
	UserUuid uuid.UUID
}

// NewHandler constructs an authorize_chat_updates handler.
func NewHandler(chatRepository domain.ChatRepository, chatMemberRepository domain.ChatMemberRepository) *Handler {
	return &Handler{
		chatRepository:       chatRepository,
		chatMemberRepository: chatMemberRepository,
	}
}

// Handle checks whether the user can receive runtime updates for the chat.
func (handler *Handler) Handle(ctx context.Context, query Query) error {
	chat, err := handler.chatRepository.FindByUuid(ctx, query.ChatUuid)
	if err != nil {
		return fmt.Errorf("finding chat by uuid %q: %w", query.ChatUuid, err)
	}

	if chat == nil {
		return ErrChatNotFound
	}

	member, err := handler.chatMemberRepository.FindByChatUuidAndUserUuid(ctx, query.ChatUuid, query.UserUuid)
	if err != nil {
		return fmt.Errorf("finding member %q in chat %q: %w", query.UserUuid, query.ChatUuid, err)
	}

	if member == nil {
		return ErrChatAccessDenied
	}

	return nil
}
