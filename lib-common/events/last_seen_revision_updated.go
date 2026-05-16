package events

import (
	"context"

	"github.com/google/uuid"
)

// LastSeenRevisionUpdated describes that a chat member advanced the last seen revision.
type LastSeenRevisionUpdated struct {
	ChatUuid         uuid.UUID `json:"chatUuid"`
	UserUuid         uuid.UUID `json:"userUuid"`
	LastSeenRevision int64     `json:"lastSeenRevision"`
}

// LastSeenRevisionUpdatedPublisher publishes last-seen-revision-updated events.
type LastSeenRevisionUpdatedPublisher interface {
	Publish(ctx context.Context, event LastSeenRevisionUpdated) error
}
