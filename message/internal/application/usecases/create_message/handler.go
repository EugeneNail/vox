package create_message

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/lib-common/validation/rules"
	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
	"github.com/samborkent/uuidv7"
)

var ErrChatNotFound = errors.New("chat not found")
var ErrChatAccessDenied = errors.New("chat access denied")

// Handler creates messages through the create_message use-case.
type Handler struct {
	messageRepository       domain.MessageRepository
	chatRepository          domain.ChatRepository
	chatMemberRepository    domain.ChatMemberRepository
	messageCreatedPublisher domain.MessageCreatedPublisher
}

// Command contains the input required to create a message.
type Command struct {
	ChatUuid uuid.UUID
	UserUuid uuid.UUID
	Text     string
}

// NewHandler constructs a create_message handler with its dependencies.
func NewHandler(messageRepository domain.MessageRepository, chatRepository domain.ChatRepository, chatMemberRepository domain.ChatMemberRepository, messageCreatedPublisher domain.MessageCreatedPublisher) *Handler {
	return &Handler{
		messageRepository:       messageRepository,
		chatRepository:          chatRepository,
		chatMemberRepository:    chatMemberRepository,
		messageCreatedPublisher: messageCreatedPublisher,
	}
}

// Handle validates input and creates a new message.
func (handler *Handler) Handle(ctx context.Context, command Command) (uuid.UUID, error) {
	text := strings.TrimSpace(command.Text)

	validator := validation.NewValidator(map[string]any{
		"chatUuid": command.ChatUuid,
		"userUuid": command.UserUuid,
		"text":     text,
	}, map[string][]rules.Rule{
		"chatUuid": {rules.Required()},
		"userUuid": {rules.Required()},
		"text":     {rules.Required(), rules.Regex(domain.MessageTextPattern), rules.Min(1), rules.Max(2000)},
	})

	if err := validator.Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return uuid.Nil, validationError
		}

		return uuid.Nil, fmt.Errorf("validating create message command: %w", err)
	}

	chat, err := handler.chatRepository.FindByUuid(ctx, command.ChatUuid)
	if err != nil {
		return uuid.Nil, fmt.Errorf("finding chat by uuid %q: %w", command.ChatUuid, err)
	}

	if chat == nil {
		return uuid.Nil, ErrChatNotFound
	}

	member, err := handler.chatMemberRepository.FindByChatUuidAndUserUuid(ctx, command.ChatUuid, command.UserUuid)
	if err != nil {
		return uuid.Nil, fmt.Errorf("finding member %q in chat %q: %w", command.UserUuid, command.ChatUuid, err)
	}

	if member == nil {
		return uuid.Nil, ErrChatAccessDenied
	}

	now := time.Now().UTC()
	message := domain.Message{
		Uuid:      uuid.UUID(uuidv7.New()),
		ChatUuid:  command.ChatUuid,
		UserUuid:  command.UserUuid,
		Text:      text,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := handler.messageRepository.Create(ctx, message); err != nil {
		return uuid.Nil, fmt.Errorf("creating message %q in chat %q: %w", message.Uuid, message.ChatUuid, err)
	}

	if err := handler.messageCreatedPublisher.Publish(ctx, domain.MessageCreatedEvent{
		MessageUuid: message.Uuid,
		ChatUuid:    message.ChatUuid,
		UserUuid:    message.UserUuid,
		Text:        message.Text,
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   message.UpdatedAt,
	}); err != nil {
		log.Printf("publishing message created event for message %q: %v", message.Uuid, err)
	}

	return message.Uuid, nil
}
