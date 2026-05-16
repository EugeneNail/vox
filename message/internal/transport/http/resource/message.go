package resource

import (
	"time"

	"github.com/google/uuid"
)

// Message represents a message resource returned by HTTP endpoints.
type Message struct {
	Uuid        uuid.UUID    `json:"uuid"`
	ChatUuid    uuid.UUID    `json:"chatUuid"`
	UserUuid    uuid.UUID    `json:"userUuid"`
	Revision    int64        `json:"revision"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}
