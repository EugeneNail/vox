package delete_message

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/lib-common/validation/rules"
	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/EugeneNail/vox/message/internal/domain/events"
	"github.com/google/uuid"
)

var ErrMessageNotFound = errors.New("message not found")
var ErrMessageAccessDenied = errors.New("message access denied")
var ErrChatNotFound = errors.New("chat not found")
var ErrChatAccessDenied = errors.New("chat access denied")

// Handler deletes messages through the delete_message use-case.
type Handler struct {
	messageRepository       domain.MessageRepository
	chatRepository          domain.ChatRepository
	chatMemberRepository    domain.ChatMemberRepository
	messageDeletedPublisher events.MessageDeletedPublisher
}

// Command contains the input required to delete a message.
type Command struct {
	MessageUuid uuid.UUID
	UserUuid    uuid.UUID
}

// NewHandler constructs a delete_message handler with its dependencies.
func NewHandler(messageRepository domain.MessageRepository, chatRepository domain.ChatRepository, chatMemberRepository domain.ChatMemberRepository, messageDeletedPublisher events.MessageDeletedPublisher) *Handler {
	return &Handler{
		messageRepository:       messageRepository,
		chatRepository:          chatRepository,
		chatMemberRepository:    chatMemberRepository,
		messageDeletedPublisher: messageDeletedPublisher,
	}
}

// Handle validates input and deletes the message.
func (handler *Handler) Handle(ctx context.Context, command Command) error {
	validator := validation.NewValidator(map[string]any{
		"messageUuid": command.MessageUuid,
		"userUuid":    command.UserUuid,
	}, map[string][]rules.Rule{
		"messageUuid": {rules.Required()},
		"userUuid":    {rules.Required()},
	})

	if err := validator.Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return validationError
		}

		return fmt.Errorf("validating delete message command: %w", err)
	}

	message, err := handler.messageRepository.FindByUuid(ctx, command.MessageUuid)
	if err != nil {
		return fmt.Errorf("finding message by uuid %q: %w", command.MessageUuid, err)
	}

	if message == nil {
		return ErrMessageNotFound
	}

	chat, err := handler.chatRepository.FindByUuid(ctx, message.ChatUuid)
	if err != nil {
		return fmt.Errorf("finding chat by uuid %q: %w", message.ChatUuid, err)
	}

	if chat == nil {
		return ErrChatNotFound
	}

	member, err := handler.chatMemberRepository.FindByChatUuidAndUserUuid(ctx, message.ChatUuid, command.UserUuid)
	if err != nil {
		return fmt.Errorf("finding member %q in chat %q: %w", command.UserUuid, message.ChatUuid, err)
	}

	if member == nil {
		return ErrChatAccessDenied
	}

	if message.UserUuid != command.UserUuid {
		return ErrMessageAccessDenied
	}

	if err := handler.messageRepository.Delete(ctx, message.Uuid); err != nil {
		return fmt.Errorf("deleting message %q in chat %q: %w", message.Uuid, message.ChatUuid, err)
	}

	if err := handler.messageDeletedPublisher.Publish(ctx, events.MessageDeleted{
		MessageUuid: message.Uuid,
		ChatUuid:    message.ChatUuid,
		UserUuid:    message.UserUuid,
	}); err != nil {
		log.Printf("publishing message deleted event for message %q: %v", message.Uuid, err)
	}

	return nil
}
