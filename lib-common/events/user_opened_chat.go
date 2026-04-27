package events

import (
	"context"

	"github.com/google/uuid"
)

// UserOpenedChat describes that a user opened a chat view in the browser.
type UserOpenedChat struct {
	UserUuid uuid.UUID `json:"userUuid"`
	ChatUuid uuid.UUID `json:"chatUuid"`
}

// UserOpenedChatPublisher publishes user-opened-chat events.
type UserOpenedChatPublisher interface {
	Publish(ctx context.Context, event UserOpenedChat) error
}
