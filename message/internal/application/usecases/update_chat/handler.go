package update_chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
)

var ErrChatNotFound = errors.New("chat not found")
var ErrChatAccessDenied = errors.New("chat access denied")

// Handler updates chat metadata through the update_chat use-case.
type Handler struct {
	chatRepository       domain.ChatRepository
	chatMemberRepository domain.ChatMemberRepository
}

// Command contains the input required to update a chat.
type Command struct {
	ChatUuid uuid.UUID
	UserUuid uuid.UUID
	Name     *string
	Avatar   *string
}

// NewHandler constructs an update_chat handler with its dependencies.
func NewHandler(chatRepository domain.ChatRepository, chatMemberRepository domain.ChatMemberRepository) *Handler {
	return &Handler{
		chatRepository:       chatRepository,
		chatMemberRepository: chatMemberRepository,
	}
}

// Handle validates input and updates the chat metadata.
func (handler *Handler) Handle(ctx context.Context, command Command) error {
	validationError := validation.NewError()

	if command.ChatUuid == uuid.Nil {
		validationError.AddViolation("chatUuid", "Required")
	}

	if command.UserUuid == uuid.Nil {
		validationError.AddViolation("userUuid", "Required")
	}

	if command.Name == nil && command.Avatar == nil {
		validationError.AddViolation("name", "At least one of name or avatar is required")
		validationError.AddViolation("avatar", "At least one of name or avatar is required")
	}

	if command.Name != nil {
		*command.Name = strings.TrimSpace(*command.Name)
		if *command.Name == "" {
			validationError.AddViolation("name", "Must not be empty")
		}
	}

	if command.Avatar != nil {
		*command.Avatar = strings.TrimSpace(*command.Avatar)
		if *command.Avatar == "" {
			validationError.AddViolation("avatar", "Must not be empty")
		}
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

	if command.Name != nil {
		chat.Name = command.Name
	}

	if command.Avatar != nil {
		chat.Avatar = command.Avatar
	}

	chat.UpdatedAt = time.Now().UTC()

	if err := handler.chatRepository.Update(ctx, *chat); err != nil {
		return fmt.Errorf("updating chat %q: %w", chat.Uuid, err)
	}

	return nil
}
