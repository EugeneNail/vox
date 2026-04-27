package events

import (
	"context"
	"time"

	"github.com/google/uuid"
)

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
