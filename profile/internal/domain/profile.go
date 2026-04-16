package domain

import (
	"time"

	"github.com/google/uuid"
)

const AvatarPattern = `^.+\.(png|jpe?g)$`

type Profile struct {
	UserUuid  uuid.UUID
	Avatar    *string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
