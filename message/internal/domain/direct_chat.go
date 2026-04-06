package domain

import (
	"time"

	"github.com/google/uuid"
)

// DirectChat represents a one-to-one chat between two members.
type DirectChat struct {
	Uuid             uuid.UUID
	FirstMemberUuid  uuid.UUID
	SecondMemberUuid uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
