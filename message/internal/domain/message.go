package domain

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	Uuid      uuid.UUID
	Text      string
	CreatedAt time.Time
	UpdatedAt time.Time

	Attachments []Attachment

	ChatUuid uuid.UUID
	UserUuid uuid.UUID
}

// MessageTextPattern allows national alphabets, combining marks, numbers,
// punctuation including all quote types, symbols, spaces, tabs, and line breaks.
const MessageTextPattern = `^[\p{L}\p{M}\p{N}\p{P}\p{S}\p{Zs}\t\r\n]*$`
