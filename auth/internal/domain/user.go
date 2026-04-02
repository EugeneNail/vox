package domain

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	Uuid      uuid.UUID
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
