package list_chat_messages

import (
	"context"
	"errors"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
)

var ErrChatNotFound = errors.New("chat not found")
var ErrChatAccessDenied = errors.New("chat access denied")

// Handler lists chat messages through the list_chat_messages use-case.
type Handler struct {
	messageRepository    domain.MessageRepository
	chatRepository       domain.ChatRepository
	chatMemberRepository domain.ChatMemberRepository
}

// Query contains the input required to list chat messages.
type Query struct {
	ChatUuid uuid.UUID
	UserUuid uuid.UUID
	Length   int
}

// NewHandler constructs a list_chat_messages handler with its dependencies.
func NewHandler(messageRepository domain.MessageRepository, chatRepository domain.ChatRepository, chatMemberRepository domain.ChatMemberRepository) *Handler {
	return &Handler{
		messageRepository:    messageRepository,
		chatRepository:       chatRepository,
		chatMemberRepository: chatMemberRepository,
	}
}

// Handle returns latest messages from a chat available to the user.
func (handler *Handler) Handle(ctx context.Context, query Query) ([]domain.Message, error) {
	chat, err := handler.chatRepository.FindByUuid(ctx, query.ChatUuid)
	if err != nil {
		return nil, fmt.Errorf("finding chat by uuid %q: %w", query.ChatUuid, err)
	}

	if chat == nil {
		return nil, ErrChatNotFound
	}

	member, err := handler.chatMemberRepository.FindByChatUuidAndUserUuid(ctx, query.ChatUuid, query.UserUuid)
	if err != nil {
		return nil, fmt.Errorf("finding member %q in chat %q: %w", query.UserUuid, query.ChatUuid, err)
	}

	if member == nil {
		return nil, ErrChatAccessDenied
	}

	messages, err := handler.messageRepository.FindLastByChatUuid(ctx, query.ChatUuid, query.Length)
	if err != nil {
		return nil, fmt.Errorf("finding last messages by chat uuid %q with length %d: %w", query.ChatUuid, query.Length, err)
	}

	return messages, nil
}
