package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/events"
	redisclient "github.com/redis/go-redis/v9"
)

const messageEditedStream = "message.edited"

// MessageEditedPublisher publishes message-edited events through Redis Streams.
type MessageEditedPublisher struct {
	client *redisclient.Client
	maxLen int64
}

// NewMessageEditedPublisher constructs a Redis-backed message-edited publisher.
func NewMessageEditedPublisher(client *redisclient.Client, maxLen int64) *MessageEditedPublisher {
	return &MessageEditedPublisher{
		client: client,
		maxLen: maxLen,
	}
}

// Publish publishes a message-edited event.
func (publisher *MessageEditedPublisher) Publish(ctx context.Context, event events.MessageEdited) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling message edited event for message %q: %w", event.MessageUuid, err)
	}

	if err := publisher.client.XAdd(ctx, &redisclient.XAddArgs{
		Stream: messageEditedStream,
		MaxLen: publisher.maxLen,
		Approx: true,
		Values: map[string]any{
			"payload": string(payload),
		},
	}).Err(); err != nil {
		return fmt.Errorf("publishing message edited event for message %q: %w", event.MessageUuid, err)
	}

	return nil
}

// Ensure MessageEditedPublisher implements the message-edited publisher contract.
var _ events.MessageEditedPublisher = (*MessageEditedPublisher)(nil)
