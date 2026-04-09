package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain/events"
	redisclient "github.com/redis/go-redis/v9"
)

const messageCreatedStream = "message.created"

// MessageCreatedPublisher publishes message-created events through Redis Streams.
type MessageCreatedPublisher struct {
	client *redisclient.Client
	maxLen int64
}

// NewMessageCreatedPublisher constructs a Redis-backed message-created publisher.
func NewMessageCreatedPublisher(client *redisclient.Client, maxLen int64) *MessageCreatedPublisher {
	return &MessageCreatedPublisher{
		client: client,
		maxLen: maxLen,
	}
}

// Publish publishes a message-created event.
func (publisher *MessageCreatedPublisher) Publish(ctx context.Context, event events.MessageCreated) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling message created event for message %q: %w", event.MessageUuid, err)
	}

	if err := publisher.client.XAdd(ctx, &redisclient.XAddArgs{
		Stream: messageCreatedStream,
		MaxLen: publisher.maxLen,
		Approx: true,
		Values: map[string]any{
			"payload": string(payload),
		},
	}).Err(); err != nil {
		return fmt.Errorf("publishing message created event for message %q: %w", event.MessageUuid, err)
	}

	return nil
}

// Ensure MessageCreatedPublisher implements the message-created publisher contract.
var _ events.MessageCreatedPublisher = (*MessageCreatedPublisher)(nil)
