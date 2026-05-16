package events

import (
	"context"

	"github.com/google/uuid"
)

// ChatRevisionUpdated describes that a chat revision advanced after a new message was created.
type ChatRevisionUpdated struct {
	ChatUuid     uuid.UUID `json:"chatUuid"`
	UserUuid     uuid.UUID `json:"userUuid"`
	MessagePiece string    `json:"messagePiece"`
	Revision     int64     `json:"revision"`
}

// ChatRevisionUpdatedPublisher publishes chat-revision-updated events.
type ChatRevisionUpdatedPublisher interface {
	Publish(ctx context.Context, event ChatRevisionUpdated) error
}
