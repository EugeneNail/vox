package events

import "github.com/google/uuid"

// Attachment describes a file attached to a message.
type Attachment struct {
	Uuid uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
}
