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
	directChatRepository domain.DirectChatRepository
}

// Query contains the input required to list chat messages.
type Query struct {
	ChatUuid uuid.UUID
	UserUuid uuid.UUID
	Length   int
}

// NewHandler constructs a list_chat_messages handler with its dependencies.
func NewHandler(messageRepository domain.MessageRepository, directChatRepository domain.DirectChatRepository) *Handler {
	return &Handler{
		messageRepository:    messageRepository,
		directChatRepository: directChatRepository,
	}
}

// Handle returns latest messages from a chat available to the user.
func (handler *Handler) Handle(ctx context.Context, query Query) ([]domain.Message, error) {
	directChat, err := handler.directChatRepository.FindByUuid(ctx, query.ChatUuid)
	if err != nil {
		return nil, fmt.Errorf("finding direct chat by uuid %q: %w", query.ChatUuid, err)
	}

	if directChat == nil {
		return nil, ErrChatNotFound
	}

	if directChat.FirstMemberUuid != query.UserUuid && directChat.SecondMemberUuid != query.UserUuid {
		return nil, ErrChatAccessDenied
	}

	messages, err := handler.messageRepository.FindLastByChatUuid(ctx, query.ChatUuid, query.Length)
	if err != nil {
		return nil, fmt.Errorf("finding last messages by chat uuid %q with length %d: %w", query.ChatUuid, query.Length, err)
	}

	return messages, nil
}
