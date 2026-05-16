package resource

import (
	"time"

	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/google/uuid"
)

// Chat represents a chat resource returned by HTTP endpoints.
type Chat struct {
	Uuid                        uuid.UUID             `json:"uuid"`
	Name                        *string               `json:"name"`
	Avatar                      *string               `json:"avatar"`
	ChatType                    domain.ChatType       `json:"chatType"`
	Revision                    int64                 `json:"revision"`
	CreatedByUserUuid           uuid.UUID             `json:"createdByUserUuid"`
	MemberUuids                 []uuid.UUID           `json:"memberUuids"`
	CurrentUserRole             domain.ChatMemberRole `json:"currentUserRole"`
	CurrentUserLastSeenRevision int64                 `json:"currentUserLastSeenRevision"`
	CreatedAt                   time.Time             `json:"createdAt"`
	UpdatedAt                   time.Time             `json:"updatedAt"`
}
