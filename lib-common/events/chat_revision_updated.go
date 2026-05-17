package events

import (
	"context"

	"github.com/google/uuid"
)

// ChatRevisionUpdated describes that a chat revision advanced after a new message was created.
type ChatRevisionUpdated struct {
	ChatUuid       uuid.UUID   `json:"chatUuid"`
	RecipientUuids []uuid.UUID `json:"recipientUuids"`
	Preview        string      `json:"preview"`
	Revision       int64       `json:"revision"`
}

// ChatRevisionUpdatedPublisher publishes chat-revision-updated events.
type ChatRevisionUpdatedPublisher interface {
	Publish(ctx context.Context, event ChatRevisionUpdated) error
}
