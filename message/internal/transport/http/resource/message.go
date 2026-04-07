package resource

import (
	"time"

	"github.com/google/uuid"
)

// Message represents a message resource returned by HTTP endpoints.
type Message struct {
	Uuid      uuid.UUID `json:"uuid"`
	ChatUuid  uuid.UUID `json:"chatUuid"`
	UserUuid  uuid.UUID `json:"userUuid"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
