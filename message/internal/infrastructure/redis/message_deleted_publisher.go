package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
	redisclient "github.com/redis/go-redis/v9"
)

const messageDeletedChannel = "message.deleted"

// MessageDeletedPublisher publishes message-deleted events through Redis Pub/Sub.
type MessageDeletedPublisher struct {
	client *redisclient.Client
}

// NewMessageDeletedPublisher constructs a Redis-backed message-deleted publisher.
func NewMessageDeletedPublisher(client *redisclient.Client) *MessageDeletedPublisher {
	return &MessageDeletedPublisher{
		client: client,
	}
}

// Publish publishes a message-deleted event.
func (publisher *MessageDeletedPublisher) Publish(ctx context.Context, event domain.MessageDeletedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling message deleted event for message %q: %w", event.MessageUuid, err)
	}

	if err := publisher.client.Publish(ctx, messageDeletedChannel, payload).Err(); err != nil {
		return fmt.Errorf("publishing message deleted event for message %q: %w", event.MessageUuid, err)
	}

	return nil
}

// Ensure MessageDeletedPublisher implements the message-deleted publisher contract.
var _ domain.MessageDeletedPublisher = (*MessageDeletedPublisher)(nil)
