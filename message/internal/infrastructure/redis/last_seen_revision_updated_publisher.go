package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/events"
	redisclient "github.com/redis/go-redis/v9"
)

// LastSeenRevisionUpdatedPublisher publishes last-seen-revision-updated events through Redis Streams.
type LastSeenRevisionUpdatedPublisher struct {
	client *redisclient.Client
	maxLen int64
}

// NewLastSeenRevisionUpdatedPublisher constructs a Redis-backed last-seen-revision-updated publisher.
func NewLastSeenRevisionUpdatedPublisher(client *redisclient.Client, maxLen int64) *LastSeenRevisionUpdatedPublisher {
	return &LastSeenRevisionUpdatedPublisher{
		client: client,
		maxLen: maxLen,
	}
}

// Publish publishes a last-seen-revision-updated event.
func (publisher *LastSeenRevisionUpdatedPublisher) Publish(ctx context.Context, event events.LastSeenRevisionUpdated) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling last seen revision updated event for chat %q and user %q: %w", event.ChatUuid, event.UserUuid, err)
	}

	if err := publisher.client.XAdd(ctx, &redisclient.XAddArgs{
		Stream: events.LastSeenRevisionUpdatedStream,
		MaxLen: publisher.maxLen,
		Approx: true,
		Values: map[string]any{
			"payload": string(payload),
		},
	}).Err(); err != nil {
		return fmt.Errorf("publishing last seen revision updated event for chat %q and user %q: %w", event.ChatUuid, event.UserUuid, err)
	}

	return nil
}

// Ensure LastSeenRevisionUpdatedPublisher implements the last-seen-revision-updated publisher contract.
var _ events.LastSeenRevisionUpdatedPublisher = (*LastSeenRevisionUpdatedPublisher)(nil)
