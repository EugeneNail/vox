package events

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MessageCreated describes a message that was persisted and can be delivered to realtime subscribers.
type MessageCreated struct {
	MessageUuid uuid.UUID `json:"messageUuid"`
	ChatUuid    uuid.UUID `json:"chatUuid"`
	UserUuid    uuid.UUID `json:"userUuid"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// MessageCreatedPublisher publishes message-created events.
type MessageCreatedPublisher interface {
	Publish(ctx context.Context, event MessageCreated) error
}

// MessageCreatedConsumer consumes message-created events.
type MessageCreatedConsumer interface {
	ListenAndConsume(ctx context.Context)
}

// MessageEdited describes a message that was edited and can be delivered to realtime subscribers.
type MessageEdited struct {
	MessageUuid uuid.UUID `json:"messageUuid"`
	ChatUuid    uuid.UUID `json:"chatUuid"`
	UserUuid    uuid.UUID `json:"userUuid"`
	Text        string    `json:"text"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// MessageEditedPublisher publishes message-edited events.
type MessageEditedPublisher interface {
	Publish(ctx context.Context, event MessageEdited) error
}

// MessageEditedConsumer consumes message-edited events.
type MessageEditedConsumer interface {
	ListenAndConsume(ctx context.Context)
}

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

// MessageDeletedConsumer consumes message-deleted events.
type MessageDeletedConsumer interface {
	ListenAndConsume(ctx context.Context)
}
