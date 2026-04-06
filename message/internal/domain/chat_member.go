package domain

import (
	"github.com/google/uuid"
	"time"
)

type ChatMember struct {
	ChatUuid   uuid.UUID
	MemberUuid uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
