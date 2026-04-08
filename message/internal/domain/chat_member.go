package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChatMemberRole string

const ChatMemberRoleMember ChatMemberRole = "member"

// ChatMember represents a user's membership in a chat.
type ChatMember struct {
	ChatUuid uuid.UUID
	UserUuid uuid.UUID
	Role     ChatMemberRole
	JoinedAt time.Time
}
