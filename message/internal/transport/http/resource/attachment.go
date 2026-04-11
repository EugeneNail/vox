package resource

import "github.com/google/uuid"

// Attachment represents a message attachment resource returned by HTTP endpoints.
type Attachment struct {
	Uuid uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
}
