package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
	redisclient "github.com/redis/go-redis/v9"
)

const messageEditedChannel = "message.edited"

// MessageEditedPublisher publishes message-edited events through Redis Pub/Sub.
type MessageEditedPublisher struct {
	client *redisclient.Client
}

// NewMessageEditedPublisher constructs a Redis-backed message-edited publisher.
func NewMessageEditedPublisher(client *redisclient.Client) *MessageEditedPublisher {
	return &MessageEditedPublisher{
		client: client,
	}
}

// Publish publishes a message-edited event.
func (publisher *MessageEditedPublisher) Publish(ctx context.Context, event domain.MessageEditedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling message edited event for message %q: %w", event.MessageUuid, err)
	}

	if err := publisher.client.Publish(ctx, messageEditedChannel, payload).Err(); err != nil {
		return fmt.Errorf("publishing message edited event for message %q: %w", event.MessageUuid, err)
	}

	return nil
}

// Ensure MessageEditedPublisher implements the message-edited publisher contract.
var _ domain.MessageEditedPublisher = (*MessageEditedPublisher)(nil)
