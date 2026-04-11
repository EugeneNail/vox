package domain

import "github.com/google/uuid"

type Attachment struct {
	Uuid uuid.UUID
	Name string

	MessageUuid uuid.UUID
}
