package open_chat

import (
	"context"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/events"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/application/usecases/authorize_chat_updates"
	"github.com/google/uuid"
)

// Handler opens a chat view and publishes a runtime subscription event.
type Handler struct {
	chatAuthorizer          *authorize_chat_updates.Handler
	userOpenedChatPublisher events.UserOpenedChatPublisher
}

// Command contains the input required to open a chat view.
type Command struct {
	ChatUuid uuid.UUID
	UserUuid uuid.UUID
}

// NewHandler constructs an open_chat handler with its dependencies.
func NewHandler(chatAuthorizer *authorize_chat_updates.Handler, userOpenedChatPublisher events.UserOpenedChatPublisher) *Handler {
	return &Handler{
		chatAuthorizer:          chatAuthorizer,
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

	if err := handler.chatAuthorizer.Handle(ctx, authorize_chat_updates.Query{
		ChatUuid: command.ChatUuid,
		UserUuid: command.UserUuid,
	}); err != nil {
		return err
	}

	if err := handler.userOpenedChatPublisher.Publish(ctx, events.UserOpenedChat{
		UserUuid: command.UserUuid,
		ChatUuid: command.ChatUuid,
	}); err != nil {
		return fmt.Errorf("publishing user opened chat event for chat %q and user %q: %w", command.ChatUuid, command.UserUuid, err)
	}

	return nil
}
