package list_chats

import (
	"context"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
)

// Handler lists chats through the list_chats use-case.
type Handler struct {
	chatRepository       domain.ChatRepository
	chatMemberRepository domain.ChatMemberRepository
}

// Query contains the input required to list chats.
type Query struct {
	UserUuid uuid.UUID
}

// Result contains chat metadata and member identifiers for transport resources.
type Result struct {
	Chat        domain.Chat
	MemberUuids []uuid.UUID
}

// NewHandler constructs a list_chats handler with its dependencies.
func NewHandler(chatRepository domain.ChatRepository, chatMemberRepository domain.ChatMemberRepository) *Handler {
	return &Handler{
		chatRepository:       chatRepository,
		chatMemberRepository: chatMemberRepository,
	}
}

// Handle returns chats available to the user.
func (handler *Handler) Handle(ctx context.Context, query Query) ([]Result, error) {
	chats, err := handler.chatRepository.FindAllByMemberUuid(ctx, query.UserUuid)
	if err != nil {
		return nil, fmt.Errorf("finding chats by user uuid %q: %w", query.UserUuid, err)
	}

	results := make([]Result, 0, len(chats))
	for _, chat := range chats {
		members, err := handler.chatMemberRepository.FindAllByChatUuid(ctx, chat.Uuid)
		if err != nil {
			return nil, fmt.Errorf("finding members by chat uuid %q: %w", chat.Uuid, err)
		}

		memberUuids := make([]uuid.UUID, 0, len(members))
		for _, member := range members {
			memberUuids = append(memberUuids, member.UserUuid)
		}

		results = append(results, Result{
			Chat:        chat,
			MemberUuids: memberUuids,
		})
	}

	return results, nil
}
