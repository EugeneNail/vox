package redis

import (
	"context"
	"encoding/json"
	"log"

	"github.com/EugeneNail/vox/message/internal/domain"
	redisclient "github.com/redis/go-redis/v9"
)

// MessageEventSubscriber listens to message events through Redis Pub/Sub.
type MessageEventSubscriber struct {
	client *redisclient.Client
}

// NewMessageEventSubscriber constructs a Redis-backed message event subscriber.
func NewMessageEventSubscriber(client *redisclient.Client) *MessageEventSubscriber {
	return &MessageEventSubscriber{
		client: client,
	}
}

// ListenMessageCreated listens for message-created events until the context is canceled.
func (subscriber *MessageEventSubscriber) ListenMessageCreated(ctx context.Context, handler func(context.Context, domain.MessageCreatedEvent) error) error {
	pubsub := subscriber.client.Subscribe(ctx, messageCreatedChannel)
	defer pubsub.Close()

	channel := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case message, ok := <-channel:
			if !ok {
				return nil
			}

			var event domain.MessageCreatedEvent
			if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
				log.Printf("unmarshalling message created event: %v", err)
				continue
			}

			if err := handler(ctx, event); err != nil {
				log.Printf("handling message created event %q: %v", event.MessageUuid, err)
				continue
			}
		}
	}
}

// Ensure MessageEventSubscriber implements the message event subscriber contract.
var _ domain.MessageEventSubscriber = (*MessageEventSubscriber)(nil)
