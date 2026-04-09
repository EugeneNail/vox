package create_chat

import (
	"context"
	"fmt"
	"time"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
	"github.com/samborkent/uuidv7"
)

// Handler creates chats through the create_chat use-case.
type Handler struct {
	chatRepository domain.ChatRepository
}

// Command contains the input required to create a chat.
type Command struct {
	CreatorUuid uuid.UUID
	MemberUuids []uuid.UUID
	Name        *string
	Avatar      *string
	IsPrivate   bool
}

// NewHandler constructs a create_chat handler with its dependencies.
func NewHandler(chatRepository domain.ChatRepository) *Handler {
	return &Handler{
		chatRepository: chatRepository,
	}
}

// Handle creates a new chat and adds all members.
func (handler *Handler) Handle(ctx context.Context, command Command) (uuid.UUID, error) {
	validationError := validation.NewError()

	if command.CreatorUuid == uuid.Nil {
		validationError.AddViolation("creatorUuid", "Required")
	}

	seenMemberUuids := make(map[uuid.UUID]struct{}, len(command.MemberUuids)+1)
	if command.CreatorUuid != uuid.Nil {
		seenMemberUuids[command.CreatorUuid] = struct{}{}
	}

	for index, memberUuid := range command.MemberUuids {
		if memberUuid == uuid.Nil {
			validationError.AddViolation(fmt.Sprintf("memberUuids.%d", index), "Must not contain empty UUID values")
			continue
		}

		if _, exists := seenMemberUuids[memberUuid]; exists {
			validationError.AddViolation(fmt.Sprintf("memberUuids.%d", index), "Must not contain duplicate UUID values")
			continue
		}

		seenMemberUuids[memberUuid] = struct{}{}
	}

	memberCount := len(command.MemberUuids) + 1
	if memberCount < 2 {
		validationError.AddViolation("memberUuids", "At least one member besides the creator is required")
	}

	if command.IsPrivate && memberCount != 2 {
		validationError.AddViolation("memberUuids", "Private chats must have exactly two members")
	}

	if len(validationError.Violations()) > 0 {
		return uuid.Nil, validationError
	}

	now := time.Now().UTC()
	chat := domain.Chat{
		Uuid:              uuid.UUID(uuidv7.New()),
		Name:              command.Name,
		Avatar:            command.Avatar,
		IsPrivate:         command.IsPrivate,
		CreatedByUserUuid: command.CreatorUuid,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	members := make([]domain.ChatMember, 0, len(command.MemberUuids)+1)
	members = append(members, domain.ChatMember{
		ChatUuid: chat.Uuid,
		UserUuid: command.CreatorUuid,
		Role:     domain.ChatMemberRoleMember,
		JoinedAt: now,
	})

	for _, memberUuid := range command.MemberUuids {
		members = append(members, domain.ChatMember{
			ChatUuid: chat.Uuid,
			UserUuid: memberUuid,
			Role:     domain.ChatMemberRoleMember,
			JoinedAt: now,
		})
	}

	if err := handler.chatRepository.CreateWithMembers(ctx, chat, members); err != nil {
		return uuid.Nil, fmt.Errorf("creating chat %q: %w", chat.Uuid, err)
	}

	return chat.Uuid, nil
}
