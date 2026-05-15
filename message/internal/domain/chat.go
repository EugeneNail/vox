package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChatType uint8

const (
	ChatTypeDirect ChatType = iota
	ChatTypeGroup
)

// Chat represents a conversation between one or more members.
type Chat struct {
	Uuid              uuid.UUID
	Name              *string
	Avatar            *string
	ChatType          ChatType
	Revision          int64
	CreatedByUserUuid uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
