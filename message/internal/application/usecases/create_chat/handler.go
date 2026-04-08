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
}

// NewHandler constructs a create_chat handler with its dependencies.
func NewHandler(chatRepository domain.ChatRepository) *Handler {
	return &Handler{
		chatRepository: chatRepository,
	}
}

// Handle creates a new chat and adds all members.
func (handler *Handler) Handle(ctx context.Context, command Command) (uuid.UUID, error) {
	if command.CreatorUuid == uuid.Nil {
		validationError := validation.NewError()
		validationError.AddViolation("creatorUuid", "Required")
		return uuid.Nil, validationError
	}

	memberUuids := normalizeMemberUuids(command.CreatorUuid, command.MemberUuids)
	if len(memberUuids) < 2 {
		validationError := validation.NewError()
		validationError.AddViolation("memberUuids", "At least one member besides the creator is required")
		return uuid.Nil, validationError
	}

	now := time.Now().UTC()
	chat := domain.Chat{
		Uuid:              uuid.UUID(uuidv7.New()),
		Name:              command.Name,
		Avatar:            command.Avatar,
		CreatedByUserUuid: command.CreatorUuid,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	members := make([]domain.ChatMember, 0, len(memberUuids))
	for _, memberUuid := range memberUuids {
		members = append(members, domain.ChatMember{
			ChatUuid: chat.Uuid,
			UserUuid: memberUuid,
			Role:     domain.ChatMemberRoleMember,
			JoinedAt: now,
		})
	}

	if err := handler.chatRepository.Create(ctx, chat, members); err != nil {
		return uuid.Nil, fmt.Errorf("creating chat %q: %w", chat.Uuid, err)
	}

	return chat.Uuid, nil
}

func normalizeMemberUuids(creatorUuid uuid.UUID, memberUuids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(memberUuids)+1)
	normalized := make([]uuid.UUID, 0, len(memberUuids)+1)

	if creatorUuid != uuid.Nil {
		seen[creatorUuid] = struct{}{}
		normalized = append(normalized, creatorUuid)
	}

	for _, memberUuid := range memberUuids {
		if memberUuid == uuid.Nil {
			continue
		}

		if _, exists := seen[memberUuid]; exists {
			continue
		}

		seen[memberUuid] = struct{}{}
		normalized = append(normalized, memberUuid)
	}

	return normalized
}
