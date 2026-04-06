package domain

import (
	"github.com/google/uuid"
	"time"
)

type Chat struct {
	Uuid      uuid.UUID
	Name      *string
	IsDirect  bool
	CreatedAt time.Time
	UpdatedAt time.Time

	ServerUuid *uuid.UUID
}
