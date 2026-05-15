package authorize_chat_access

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

// Handler authorizes access to a chat.
type Handler struct {
	chatRepository       domain.ChatRepository
	chatMemberRepository domain.ChatMemberRepository
}

// Command contains the input required to authorize chat access.
type Command struct {
	ChatUuid uuid.UUID
	UserUuid uuid.UUID
}

// NewHandler constructs an authorize_chat_access handler with its dependencies.
func NewHandler(chatRepository domain.ChatRepository, chatMemberRepository domain.ChatMemberRepository) *Handler {
	return &Handler{
		chatRepository:       chatRepository,
		chatMemberRepository: chatMemberRepository,
	}
}

// Handle validates input and authorizes access to the chat.
func (handler *Handler) Handle(ctx context.Context, command Command) error {
	validationError := validation.NewError()
	if command.ChatUuid == uuid.Nil {
		validationError.AddViolation("chatUuid", "Required")
	}

	if command.UserUuid == uuid.Nil {
		validationError.AddViolation("userUuid", "Required")
	}

	if len(validationError.Violations()) > 0 {
		return validationError
	}

	chat, err := handler.chatRepository.FindByUuid(ctx, command.ChatUuid)
	if err != nil {
		return fmt.Errorf("finding chat by uuid %q: %w", command.ChatUuid, err)
	}

	if chat == nil {
		return ErrChatNotFound
	}

	member, err := handler.chatMemberRepository.FindByChatUuidAndUserUuid(ctx, command.ChatUuid, command.UserUuid)
	if err != nil {
		return fmt.Errorf("finding member %q in chat %q: %w", command.UserUuid, command.ChatUuid, err)
	}

	if member == nil {
		return ErrChatAccessDenied
	}

	return nil
}
