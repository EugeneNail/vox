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

// MessageEventPublisher publishes message domain events.
type MessageEventPublisher interface {
	PublishMessageCreated(ctx context.Context, event MessageCreatedEvent) error
}

// MessageEventSubscriber listens to message domain events.
type MessageEventSubscriber interface {
	ListenMessageCreated(ctx context.Context, handler func(context.Context, MessageCreatedEvent) error) error
}
