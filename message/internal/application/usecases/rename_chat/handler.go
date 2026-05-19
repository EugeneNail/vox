package rename_chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/lib-common/validation/rules"
	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
)

var ErrChatNotFound = errors.New("chat not found")
var ErrChatAccessDenied = errors.New("chat access denied")

// Handler renames chats through the rename_chat use-case.
type Handler struct {
	chatRepository       domain.ChatRepository
	chatMemberRepository domain.ChatMemberRepository
}

// Command contains the input required to rename a chat.
type Command struct {
	ChatUuid uuid.UUID
	UserUuid uuid.UUID
	Name     *string
}

// NewHandler constructs a rename_chat handler with its dependencies.
func NewHandler(chatRepository domain.ChatRepository, chatMemberRepository domain.ChatMemberRepository) *Handler {
	return &Handler{
		chatRepository:       chatRepository,
		chatMemberRepository: chatMemberRepository,
	}
}

// Handle validates input and renames the chat.
func (handler *Handler) Handle(ctx context.Context, command Command) error {
	name := ""
	if command.Name != nil {
		name = strings.TrimSpace(*command.Name)
	}

	validator := validation.NewValidator(map[string]any{
		"chatUuid": command.ChatUuid,
		"userUuid": command.UserUuid,
		"name":     name,
	}, map[string][]rules.Rule{
		"chatUuid": {rules.Required()},
		"userUuid": {rules.Required()},
		"name":     {rules.Required(), rules.Min(3), rules.Max(128), rules.Regex(rules.SlugWithSpacesPattern)},
	})

	if err := validator.Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return validationError
		}
		return fmt.Errorf("validating rename chat command: %w", err)
	}

	chat, err := handler.chatRepository.FindByUuid(ctx, command.ChatUuid)
	if err != nil {
		return fmt.Errorf("finding chat by uuid %q: %w", command.ChatUuid, err)
	}

	if chat == nil {
		return ErrChatNotFound
	}

	if chat.CreatedByUserUuid != command.UserUuid {
		return ErrChatAccessDenied
	}

	member, err := handler.chatMemberRepository.FindByChatUuidAndUserUuid(ctx, command.ChatUuid, command.UserUuid)
	if err != nil {
		return fmt.Errorf("finding member %q in chat %q: %w", command.UserUuid, command.ChatUuid, err)
	}

	if member == nil {
		return ErrChatAccessDenied
	}

	chat.Name = &name
	chat.UpdatedAt = time.Now().UTC()

	if err := handler.chatRepository.Update(ctx, *chat); err != nil {
		return fmt.Errorf("updating chat %q: %w", chat.Uuid, err)
	}

	return nil
}
