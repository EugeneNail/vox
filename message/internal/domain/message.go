package domain

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	Uuid      uuid.UUID
	ChatUuid  uuid.UUID
	UserUuid  uuid.UUID
	Text      string
	CreatedAt time.Time
}
