package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/events"
	redisclient "github.com/redis/go-redis/v9"
)

// ChatRevisionUpdatedPublisher publishes chat-revision-updated events through Redis Streams.
type ChatRevisionUpdatedPublisher struct {
	client *redisclient.Client
	maxLen int64
}

// NewChatRevisionUpdatedPublisher constructs a Redis-backed chat-revision-updated publisher.
func NewChatRevisionUpdatedPublisher(client *redisclient.Client, maxLen int64) *ChatRevisionUpdatedPublisher {
	return &ChatRevisionUpdatedPublisher{
		client: client,
		maxLen: maxLen,
	}
}

// Publish publishes a chat-revision-updated event.
func (publisher *ChatRevisionUpdatedPublisher) Publish(ctx context.Context, event events.ChatRevisionUpdated) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling chat revision updated event for chat %q: %w", event.ChatUuid, err)
	}

	if err := publisher.client.XAdd(ctx, &redisclient.XAddArgs{
		Stream: events.ChatRevisionUpdatedStream,
		MaxLen: publisher.maxLen,
		Approx: true,
		Values: map[string]any{
			"payload": string(payload),
		},
	}).Err(); err != nil {
		return fmt.Errorf("publishing chat revision updated event for chat %q: %w", event.ChatUuid, err)
	}

	return nil
}

// Ensure ChatRevisionUpdatedPublisher implements the chat-revision-updated publisher contract.
var _ events.ChatRevisionUpdatedPublisher = (*ChatRevisionUpdatedPublisher)(nil)
