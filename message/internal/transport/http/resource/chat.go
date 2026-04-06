package resource

import "github.com/google/uuid"

type Chat struct {
	Uuid uuid.UUID `json:"uuid"`
	Name *string   `json:"name"`
}
