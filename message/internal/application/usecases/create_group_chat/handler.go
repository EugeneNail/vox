package create_group_chat

import (
	"context"
	"fmt"
	"time"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
	"github.com/samborkent/uuidv7"
)

// Handler creates group chats through the create_group_chat use-case.
type Handler struct {
	chatRepository domain.ChatRepository
}

// Command contains the input required to create a group chat.
type Command struct {
	CreatorUuid uuid.UUID
	MemberUuids []uuid.UUID
	Name        *string
	Avatar      *string
}

// NewHandler constructs a create_group_chat handler with its dependencies.
func NewHandler(chatRepository domain.ChatRepository) *Handler {
	return &Handler{
		chatRepository: chatRepository,
	}
}

// Handle creates a group chat and returns its UUID.
func (handler *Handler) Handle(ctx context.Context, command Command) (uuid.UUID, error) {
	validationError := validation.NewError()

	if command.CreatorUuid == uuid.Nil {
		validationError.AddViolation("creatorUuid", "Required")
	}

	membersToAdd := make(map[uuid.UUID]struct{}, len(command.MemberUuids)+1)
	if command.CreatorUuid != uuid.Nil {
		membersToAdd[command.CreatorUuid] = struct{}{}
	}

	for index, memberUuid := range command.MemberUuids {
		if memberUuid == uuid.Nil {
			validationError.AddViolation(fmt.Sprintf("memberUuids.%d", index), "Must not contain empty UUID values")
			continue
		}

		if _, exists := membersToAdd[memberUuid]; exists {
			validationError.AddViolation(fmt.Sprintf("memberUuids.%d", index), "Must not contain duplicate UUID values")
			continue
		}

		membersToAdd[memberUuid] = struct{}{}
	}

	if len(validationError.Violations()) > 0 {
		return uuid.Nil, validationError
	}

	now := time.Now().UTC()
	chat := domain.Chat{
		Uuid:              uuid.UUID(uuidv7.New()),
		Name:              command.Name,
		Avatar:            command.Avatar,
		ChatType:          domain.ChatTypeGroup,
		Revision:          0,
		CreatedByUserUuid: command.CreatorUuid,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	members := make([]domain.ChatMember, 0, len(command.MemberUuids)+1)
	members = append(members, domain.ChatMember{
		ChatUuid:         chat.Uuid,
		UserUuid:         command.CreatorUuid,
		Role:             domain.ChatMemberRoleOwner,
		LastSeenRevision: 0,
		JoinedAt:         now,
	})

	for _, memberUuid := range command.MemberUuids {
		members = append(members, domain.ChatMember{
			ChatUuid:         chat.Uuid,
			UserUuid:         memberUuid,
			Role:             domain.ChatMemberRoleMember,
			LastSeenRevision: 0,
			JoinedAt:         now,
		})
	}

	if err := handler.chatRepository.CreateWithMembers(ctx, chat, members); err != nil {
		return uuid.Nil, fmt.Errorf("creating group chat %q: %w", chat.Uuid, err)
	}

	return chat.Uuid, nil
}
