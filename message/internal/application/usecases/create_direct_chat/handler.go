package create_direct_chat

import (
	"context"
	"fmt"
	"time"

	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
	"github.com/samborkent/uuidv7"
)

// Handler creates direct chats through the create_direct_chat use-case.
type Handler struct {
	chatRepository       domain.ChatRepository
	chatMemberRepository domain.ChatMemberRepository
}

// Command contains the input required to create a direct chat.
type Command struct {
	CompanionUuid uuid.UUID
	CreatorUuid   uuid.UUID
}

// NewHandler constructs a create_direct_chat handler with its dependencies.
func NewHandler(chatRepository domain.ChatRepository, chatMemberRepository domain.ChatMemberRepository) *Handler {
	return &Handler{
		chatRepository:       chatRepository,
		chatMemberRepository: chatMemberRepository,
	}
}

// Handle creates a new direct chat and adds both members.
func (handler *Handler) Handle(ctx context.Context, command Command) (uuid.UUID, error) {
	// TODO: Validate that companion user exists through auth gRPC before creating a direct chat.
	now := time.Now().UTC()
	chat := domain.Chat{
		Uuid:      uuid.UUID(uuidv7.New()),
		IsDirect:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := handler.chatRepository.Create(ctx, chat); err != nil {
		return uuid.Nil, fmt.Errorf("creating direct chat %q: %w", chat.Uuid, err)
	}

	memberUuids := []uuid.UUID{command.CreatorUuid, command.CompanionUuid}
	for _, memberUuid := range memberUuids {
		chatMember := domain.ChatMember{
			ChatUuid:   chat.Uuid,
			MemberUuid: memberUuid,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		if err := handler.chatMemberRepository.Create(ctx, chatMember); err != nil {
			return uuid.Nil, fmt.Errorf("creating direct chat member %q in chat %q: %w", memberUuid, chat.Uuid, err)
		}
	}

	return chat.Uuid, nil
}
