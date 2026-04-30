package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChatMemberRole string

const (
	ChatMemberRoleOwner  ChatMemberRole = "owner"
	ChatMemberRoleAdmin  ChatMemberRole = "admin"
	ChatMemberRoleMember ChatMemberRole = "member"
)

// ChatMember represents a user's membership in a chat.
type ChatMember struct {
	ChatUuid uuid.UUID
	UserUuid uuid.UUID
	Role     ChatMemberRole
	JoinedAt time.Time
}
