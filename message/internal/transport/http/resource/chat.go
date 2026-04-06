package resource

import "github.com/google/uuid"

// DirectChat represents a direct chat resource returned by HTTP endpoints.
type DirectChat struct {
	Uuid             uuid.UUID `json:"uuid"`
	FirstMemberUuid  uuid.UUID `json:"firstMemberUuid"`
	SecondMemberUuid uuid.UUID `json:"secondMemberUuid"`
}
