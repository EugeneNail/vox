package domain

import (
	"time"

	"github.com/google/uuid"
)

// Chat represents a conversation between one or more members.
type Chat struct {
	Uuid              uuid.UUID
	Name              *string
	Avatar            *string
	IsPrivate         bool
	CreatedByUserUuid uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
