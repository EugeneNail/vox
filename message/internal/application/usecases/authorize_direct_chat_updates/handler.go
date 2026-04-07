package authorize_direct_chat_updates

import (
	"context"
	"errors"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
)

var ErrDirectChatNotFound = errors.New("direct chat not found")
var ErrDirectChatAccessDenied = errors.New("direct chat access denied")

// Handler authorizes runtime direct-chat update subscriptions.
type Handler struct {
	directChatRepository domain.DirectChatRepository
}

// Query contains the input required to authorize direct-chat realtime updates.
type Query struct {
	DirectChatUuid uuid.UUID
	UserUuid       uuid.UUID
}

// NewHandler constructs an authorize_direct_chat_updates handler.
func NewHandler(directChatRepository domain.DirectChatRepository) *Handler {
	return &Handler{
		directChatRepository: directChatRepository,
	}
}

// Handle checks whether the user can receive runtime updates for the direct chat.
func (handler *Handler) Handle(ctx context.Context, query Query) error {
	directChat, err := handler.directChatRepository.FindByUuid(ctx, query.DirectChatUuid)
	if err != nil {
		return fmt.Errorf("finding direct chat by uuid %q: %w", query.DirectChatUuid, err)
	}

	if directChat == nil {
		return ErrDirectChatNotFound
	}

	if directChat.FirstMemberUuid != query.UserUuid && directChat.SecondMemberUuid != query.UserUuid {
		return ErrDirectChatAccessDenied
	}

	return nil
}
