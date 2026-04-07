package redis

import (
	"context"
	"encoding/json"
	"log"

	"github.com/EugeneNail/vox/message/internal/domain"
	redisclient "github.com/redis/go-redis/v9"
)

// MessageCreatedListener listens to message-created events through Redis Pub/Sub.
type MessageCreatedListener struct {
	client *redisclient.Client
}

// NewMessageCreatedListener constructs a Redis-backed message-created listener.
func NewMessageCreatedListener(client *redisclient.Client) *MessageCreatedListener {
	return &MessageCreatedListener{
		client: client,
	}
}

// Listen listens for message-created events until the context is canceled.
func (listener *MessageCreatedListener) Listen(ctx context.Context, handler func(context.Context, domain.MessageCreatedEvent) error) error {
	pubsub := listener.client.Subscribe(ctx, messageCreatedChannel)
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

// Ensure MessageCreatedListener implements the message-created listener contract.
var _ domain.MessageCreatedListener = (*MessageCreatedListener)(nil)
