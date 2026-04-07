package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
	redisclient "github.com/redis/go-redis/v9"
)

const messageCreatedChannel = "message.created"

// MessageEventPublisher publishes message events through Redis Pub/Sub.
type MessageEventPublisher struct {
	client *redisclient.Client
}

// NewMessageEventPublisher constructs a Redis-backed message event publisher.
func NewMessageEventPublisher(client *redisclient.Client) *MessageEventPublisher {
	return &MessageEventPublisher{
		client: client,
	}
}

// PublishMessageCreated publishes a message-created event.
func (publisher *MessageEventPublisher) PublishMessageCreated(ctx context.Context, event domain.MessageCreatedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling message created event for message %q: %w", event.MessageUuid, err)
	}

	if err := publisher.client.Publish(ctx, messageCreatedChannel, payload).Err(); err != nil {
		return fmt.Errorf("publishing message created event for message %q: %w", event.MessageUuid, err)
	}

	return nil
}

// Ensure MessageEventPublisher implements the message event publisher contract.
var _ domain.MessageEventPublisher = (*MessageEventPublisher)(nil)
