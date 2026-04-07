package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MessageCreatedEvent describes a message that was persisted and can be delivered to realtime subscribers.
type MessageCreatedEvent struct {
	MessageUuid uuid.UUID `json:"messageUuid"`
	ChatUuid    uuid.UUID `json:"chatUuid"`
	UserUuid    uuid.UUID `json:"userUuid"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// MessageCreatedPublisher publishes message-created events.
type MessageCreatedPublisher interface {
	Publish(ctx context.Context, event MessageCreatedEvent) error
}

// MessageCreatedConsumer consumes message-created events.
type MessageCreatedConsumer interface {
	Start(ctx context.Context) error
}
