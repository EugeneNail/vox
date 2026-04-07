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

// MessageEditedEvent describes a message that was edited and can be delivered to realtime subscribers.
type MessageEditedEvent struct {
	MessageUuid uuid.UUID `json:"messageUuid"`
	ChatUuid    uuid.UUID `json:"chatUuid"`
	UserUuid    uuid.UUID `json:"userUuid"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// MessageEditedPublisher publishes message-edited events.
type MessageEditedPublisher interface {
	Publish(ctx context.Context, event MessageEditedEvent) error
}

// MessageEditedConsumer consumes message-edited events.
type MessageEditedConsumer interface {
	Start(ctx context.Context) error
}
