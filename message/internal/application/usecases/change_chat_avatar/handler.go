package change_chat_avatar

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

// Handler changes chat avatar through the change_chat_avatar use-case.
type Handler struct {
	chatRepository       domain.ChatRepository
	chatMemberRepository domain.ChatMemberRepository
}

// Command contains the input required to change a chat avatar.
type Command struct {
	ChatUuid uuid.UUID
	UserUuid uuid.UUID
	Avatar   *string
}

// NewHandler constructs a change_chat_avatar handler with its dependencies.
func NewHandler(chatRepository domain.ChatRepository, chatMemberRepository domain.ChatMemberRepository) *Handler {
	return &Handler{
		chatRepository:       chatRepository,
		chatMemberRepository: chatMemberRepository,
	}
}

// Handle validates input and changes the chat avatar.
func (handler *Handler) Handle(ctx context.Context, command Command) error {
	validationError := validation.NewError()

	if command.ChatUuid == uuid.Nil {
		validationError.AddViolation("chatUuid", "Required")
	}

	if command.UserUuid == uuid.Nil {
		validationError.AddViolation("userUuid", "Required")
	}

	avatar := ""
	if command.Avatar != nil {
		avatar = strings.TrimSpace(*command.Avatar)
	}

	validator := validation.NewValidator(map[string]any{
		"chatUuid": command.ChatUuid,
		"userUuid": command.UserUuid,
		"avatar":   avatar,
	}, map[string][]rules.Rule{
		"chatUuid": {rules.Required()},
		"userUuid": {rules.Required()},
		"avatar":   {rules.Required(), rules.Regex(`^.+\.[a-z]{3,4}$`)},
	})

	if err := validator.Validate(); err != nil {
		var currentValidationError validation.Error
		if errors.As(err, &currentValidationError) {
			mergeValidationErrors(validationError, currentValidationError)
		} else {
			return fmt.Errorf("validating change chat avatar command: %w", err)
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

	chat.Avatar = &avatar
	chat.UpdatedAt = time.Now().UTC()

	if err := handler.chatRepository.Update(ctx, *chat); err != nil {
		return fmt.Errorf("updating chat %q: %w", chat.Uuid, err)
	}

	return nil
}

func mergeValidationErrors(target validation.Error, source validation.Error) {
	for field, message := range source.Violations() {
		target.AddViolation(field, message)
	}
}
