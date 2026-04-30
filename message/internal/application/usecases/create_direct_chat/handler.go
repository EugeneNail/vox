package create_direct_chat

import (
	"context"
	"fmt"
	"time"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
	"github.com/samborkent/uuidv7"
)

// Handler creates direct chats through the create_direct_chat use-case.
type Handler struct {
	chatRepository domain.ChatRepository
}

// Command contains the input required to create a direct chat.
type Command struct {
	CreatorUuid      uuid.UUID
	InterlocutorUuid uuid.UUID
}

// NewHandler constructs a create_direct_chat handler with its dependencies.
func NewHandler(chatRepository domain.ChatRepository) *Handler {
	return &Handler{
		chatRepository: chatRepository,
	}
}

// Handle creates a direct chat and returns its UUID.
func (handler *Handler) Handle(ctx context.Context, command Command) (uuid.UUID, error) {
	validationError := validation.NewError()

	if command.CreatorUuid == uuid.Nil {
		validationError.AddViolation("creatorUuid", "Required")
	}

	if command.InterlocutorUuid == uuid.Nil {
		validationError.AddViolation("interlocutorUuid", "Required")
	}

	if command.InterlocutorUuid != uuid.Nil && command.InterlocutorUuid == command.CreatorUuid {
		validationError.AddViolation("interlocutorUuid", "Must not match the creator UUID")
	}

	if len(validationError.Violations()) > 0 {
		return uuid.Nil, validationError
	}

	// TODO: validate that interlocutorUuid refers to an existing user by querying auth before creating the chat.
	existingChat, err := handler.chatRepository.FindDirectByMemberUuids(ctx, command.CreatorUuid, command.InterlocutorUuid)
	if err != nil {
		return uuid.Nil, fmt.Errorf(
			"finding existing direct chat for members %q and %q: %w",
			command.CreatorUuid,
			command.InterlocutorUuid,
			err,
		)
	}

	if existingChat != nil {
		return existingChat.Uuid, nil
	}

	now := time.Now().UTC()
	chat := domain.Chat{
		Uuid:              uuid.UUID(uuidv7.New()),
		Name:              nil,
		Avatar:            nil,
		ChatType:          domain.ChatTypeDirect,
		CreatedByUserUuid: command.CreatorUuid,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	members := []domain.ChatMember{
		{
			ChatUuid: chat.Uuid,
			UserUuid: command.CreatorUuid,
			Role:     domain.ChatMemberRoleMember,
			JoinedAt: now,
		},
		{
			ChatUuid: chat.Uuid,
			UserUuid: command.InterlocutorUuid,
			Role:     domain.ChatMemberRoleMember,
			JoinedAt: now,
		},
	}

	if err := handler.chatRepository.CreateWithMembers(ctx, chat, members); err != nil {
		return uuid.Nil, fmt.Errorf("creating direct chat %q: %w", chat.Uuid, err)
	}

	return chat.Uuid, nil
}
