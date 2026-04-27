package events

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Attachment describes a file attached to a message.
type Attachment struct {
	Uuid uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
}

// MessageCreated describes a message that was persisted and can be delivered to realtime subscribers.
type MessageCreated struct {
	MessageUuid uuid.UUID    `json:"messageUuid"`
	ChatUuid    uuid.UUID    `json:"chatUuid"`
	UserUuid    uuid.UUID    `json:"userUuid"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

// MessageCreatedPublisher publishes message-created events.
type MessageCreatedPublisher interface {
	Publish(ctx context.Context, event MessageCreated) error
}

// MessageEdited describes a message that was edited and can be delivered to realtime subscribers.
type MessageEdited struct {
	MessageUuid uuid.UUID    `json:"messageUuid"`
	ChatUuid    uuid.UUID    `json:"chatUuid"`
	UserUuid    uuid.UUID    `json:"userUuid"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

// MessageEditedPublisher publishes message-edited events.
type MessageEditedPublisher interface {
	Publish(ctx context.Context, event MessageEdited) error
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

// ChatViewOpened describes that a user opened a chat view in the browser.
type ChatViewOpened struct {
	UserUuid uuid.UUID `json:"userUuid"`
	ChatUuid uuid.UUID `json:"chatUuid"`
}

// ChatViewOpenedPublisher publishes chat-view-opened events.
type ChatViewOpenedPublisher interface {
	Publish(ctx context.Context, event ChatViewOpened) error
}

// ChatViewClosed describes that a user closed a chat view in the browser.
type ChatViewClosed struct {
	UserUuid uuid.UUID `json:"userUuid"`
	ChatUuid uuid.UUID `json:"chatUuid"`
}
