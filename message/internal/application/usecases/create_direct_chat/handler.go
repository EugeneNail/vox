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
	directChatRepository domain.DirectChatRepository
}

// Command contains the input required to create a direct chat.
type Command struct {
	CompanionUuid uuid.UUID
	CreatorUuid   uuid.UUID
}

// NewHandler constructs a create_direct_chat handler with its dependencies.
func NewHandler(directChatRepository domain.DirectChatRepository) *Handler {
	return &Handler{
		directChatRepository: directChatRepository,
	}
}

// Handle creates a new direct chat and adds both members.
func (handler *Handler) Handle(ctx context.Context, command Command) (uuid.UUID, error) {
	// TODO: Validate that companion user exists through auth gRPC before creating a direct chat.
	now := time.Now().UTC()
	chat := domain.DirectChat{
		Uuid:             uuid.UUID(uuidv7.New()),
		FirstMemberUuid:  command.CreatorUuid,
		SecondMemberUuid: command.CompanionUuid,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := handler.directChatRepository.Create(ctx, chat); err != nil {
		return uuid.Nil, fmt.Errorf("creating direct chat %q: %w", chat.Uuid, err)
	}

	return chat.Uuid, nil
}
