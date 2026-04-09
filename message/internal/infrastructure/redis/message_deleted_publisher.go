package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
	redisclient "github.com/redis/go-redis/v9"
)

const messageDeletedStream = "message.deleted"

// MessageDeletedPublisher publishes message-deleted events through Redis Streams.
type MessageDeletedPublisher struct {
	client *redisclient.Client
	maxLen int64
}

// NewMessageDeletedPublisher constructs a Redis-backed message-deleted publisher.
func NewMessageDeletedPublisher(client *redisclient.Client, maxLen int64) *MessageDeletedPublisher {
	return &MessageDeletedPublisher{
		client: client,
		maxLen: maxLen,
	}
}

// Publish publishes a message-deleted event.
func (publisher *MessageDeletedPublisher) Publish(ctx context.Context, event domain.MessageDeletedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling message deleted event for message %q: %w", event.MessageUuid, err)
	}

	if err := publisher.client.XAdd(ctx, &redisclient.XAddArgs{
		Stream: messageDeletedStream,
		MaxLen: publisher.maxLen,
		Approx: true,
		Values: map[string]any{
			"payload": string(payload),
		},
	}).Err(); err != nil {
		return fmt.Errorf("publishing message deleted event for message %q: %w", event.MessageUuid, err)
	}

	return nil
}

// Ensure MessageDeletedPublisher implements the message-deleted publisher contract.
var _ domain.MessageDeletedPublisher = (*MessageDeletedPublisher)(nil)
