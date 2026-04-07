package redis

import (
	"context"
	"encoding/json"
	"log"

	"github.com/EugeneNail/vox/message/internal/domain"
	redisclient "github.com/redis/go-redis/v9"
)

type MessageCreatedHandler func(context.Context, domain.MessageCreatedEvent) error

// MessageCreatedConsumer consumes message-created events through Redis Pub/Sub.
type MessageCreatedConsumer struct {
	client   *redisclient.Client
	handlers []MessageCreatedHandler
}

// NewMessageCreatedConsumer constructs a Redis-backed message-created consumer.
func NewMessageCreatedConsumer(client *redisclient.Client, handlers ...MessageCreatedHandler) *MessageCreatedConsumer {
	return &MessageCreatedConsumer{
		client:   client,
		handlers: handlers,
	}
}

// Start consumes message-created events until the context is canceled.
func (consumer *MessageCreatedConsumer) Start(ctx context.Context) error {
	pubsub := consumer.client.Subscribe(ctx, messageCreatedChannel)
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

			for _, handler := range consumer.handlers {
				if err := handler(ctx, event); err != nil {
					log.Printf("handling message created event %q: %v", event.MessageUuid, err)
					continue
				}
			}
		}
	}
}

// Ensure MessageCreatedConsumer implements the message-created consumer contract.
var _ domain.MessageCreatedConsumer = (*MessageCreatedConsumer)(nil)
