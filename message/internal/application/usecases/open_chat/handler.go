package open_chat

import (
	"context"
	"errors"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/events"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
)

var ErrChatNotFound = errors.New("chat not found")
var ErrChatAccessDenied = errors.New("chat access denied")

// Handler opens a chat view and publishes a runtime subscription event.
type Handler struct {
	chatRepository          domain.ChatRepository
	chatMemberRepository    domain.ChatMemberRepository
	userOpenedChatPublisher events.UserOpenedChatPublisher
}

// Command contains the input required to open a chat view.
type Command struct {
	ChatUuid uuid.UUID
	UserUuid uuid.UUID
}

// NewHandler constructs an open_chat handler with its dependencies.
func NewHandler(chatRepository domain.ChatRepository, chatMemberRepository domain.ChatMemberRepository, userOpenedChatPublisher events.UserOpenedChatPublisher) *Handler {
	return &Handler{
		chatRepository:          chatRepository,
		chatMemberRepository:    chatMemberRepository,
		userOpenedChatPublisher: userOpenedChatPublisher,
	}
}

// Handle validates input, authorizes access and publishes a user-opened-chat event.
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

	if err := handler.userOpenedChatPublisher.Publish(ctx, events.UserOpenedChat{
		UserUuid: command.UserUuid,
		ChatUuid: command.ChatUuid,
	}); err != nil {
		return fmt.Errorf("publishing user opened chat event for chat %q and user %q: %w", command.ChatUuid, command.UserUuid, err)
	}

	return nil
}
