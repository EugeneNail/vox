package events

import (
	"time"

	"github.com/google/uuid"
)

// UserCreated is emitted after auth successfully creates a user.
type UserCreated struct {
	UserUuid  uuid.UUID `json:"userUuid"`
	Name      string    `json:"name"`
	Nickname  string    `json:"nickname"`
	CreatedAt time.Time `json:"createdAt"`
}
