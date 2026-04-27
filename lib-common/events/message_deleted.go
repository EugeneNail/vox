package events

import (
	"context"

	"github.com/google/uuid"
)

// MessageDeleted describes a message that was deleted and can be delivered to realtime subscribers.
type MessageDeleted struct {
	MessageUuid uuid.UUID `json:"messageUuid"`
	ChatUuid    uuid.UUID `json:"chatUuid"`
	UserUuid    uuid.UUID `json:"userUuid"`
}

// MessageDeletedPublisher publishes message-deleted events.
type MessageDeletedPublisher interface {
	Publish(ctx context.Context, event MessageDeleted) error
}
