package list_chat_messages

import (
	"context"
	"errors"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/validation"
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
	Revision int64
}

// NewHandler constructs a list_chat_messages handler with its dependencies.
func NewHandler(messageRepository domain.MessageRepository, chatRepository domain.ChatRepository, chatMemberRepository domain.ChatMemberRepository) *Handler {
	return &Handler{
		messageRepository:    messageRepository,
		chatRepository:       chatRepository,
		chatMemberRepository: chatMemberRepository,
	}
}

// Handle returns chat messages with revision greater than or equal to the requested revision.
func (handler *Handler) Handle(ctx context.Context, query Query) ([]domain.Message, error) {
	validationError := validation.NewError()
	if query.ChatUuid == uuid.Nil {
		validationError.AddViolation("chatUuid", "Required")
	}

	if query.UserUuid == uuid.Nil {
		validationError.AddViolation("userUuid", "Required")
	}

	if query.Revision < 0 {
		validationError.AddViolation("revision", "Must be greater than or equal to zero")
	}

	if len(validationError.Violations()) > 0 {
		return nil, validationError
	}

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

	validationError = validation.NewError()
	if query.Revision > chat.Revision {
		validationError.AddViolation("revision", fmt.Sprintf("Must not exceed chat revision %d", chat.Revision))
	}

	if len(validationError.Violations()) > 0 {
		return nil, validationError
	}

	messages, err := handler.messageRepository.ListFromRevision(ctx, query.ChatUuid, query.Revision)
	if err != nil {
		return nil, fmt.Errorf("listing messages by chat uuid %q from revision %d: %w", query.ChatUuid, query.Revision, err)
	}

	return messages, nil
}
