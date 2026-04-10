package resource

import "github.com/google/uuid"

// Profile represents a public profile resource returned by HTTP endpoints.
type Profile struct {
	UserUuid uuid.UUID `json:"userUuid"`
	Avatar   *string   `json:"avatar"`
	Name     string    `json:"name"`
	Nickname string    `json:"nickname"`
}
