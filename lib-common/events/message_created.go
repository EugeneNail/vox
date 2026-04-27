package events

import (
	"context"
	"time"

	"github.com/google/uuid"
)

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
