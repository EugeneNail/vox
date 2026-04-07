package redis

import (
	"context"
	"encoding/json"
	"log"

	"github.com/EugeneNail/vox/message/internal/domain"
	redisclient "github.com/redis/go-redis/v9"
)

type MessageEditedHandler func(context.Context, domain.MessageEditedEvent) error

// MessageEditedConsumer consumes message-edited events through Redis Pub/Sub.
type MessageEditedConsumer struct {
	client   *redisclient.Client
	handlers []MessageEditedHandler
}

// NewMessageEditedConsumer constructs a Redis-backed message-edited consumer.
func NewMessageEditedConsumer(client *redisclient.Client, handlers ...MessageEditedHandler) *MessageEditedConsumer {
	return &MessageEditedConsumer{
		client:   client,
		handlers: handlers,
	}
}

// Start consumes message-edited events until the context is canceled.
func (consumer *MessageEditedConsumer) Start(ctx context.Context) error {
	pubsub := consumer.client.Subscribe(ctx, messageEditedChannel)
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

			var event domain.MessageEditedEvent
			if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
				log.Printf("unmarshalling message edited event: %v", err)
				continue
			}

			for _, handler := range consumer.handlers {
				if err := handler(ctx, event); err != nil {
					log.Printf("handling message edited event %q: %v", event.MessageUuid, err)
					continue
				}
			}
		}
	}
}

// Ensure MessageEditedConsumer implements the message-edited consumer contract.
var _ domain.MessageEditedConsumer = (*MessageEditedConsumer)(nil)
