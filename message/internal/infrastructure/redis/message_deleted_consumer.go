package redis

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/EugeneNail/vox/message/internal/domain"
	redisclient "github.com/redis/go-redis/v9"
)

type MessageDeletedHandler func(context.Context, domain.MessageDeletedEvent) error

// MessageDeletedConsumer consumes message-deleted events through Redis Pub/Sub.
type MessageDeletedConsumer struct {
	client   *redisclient.Client
	handlers []MessageDeletedHandler
}

// NewMessageDeletedConsumer constructs a Redis-backed message-deleted consumer.
func NewMessageDeletedConsumer(client *redisclient.Client, handlers ...MessageDeletedHandler) *MessageDeletedConsumer {
	return &MessageDeletedConsumer{
		client:   client,
		handlers: handlers,
	}
}

// ListenAndConsume starts message-deleted consumption in a goroutine and logs unexpected errors.
func (consumer *MessageDeletedConsumer) ListenAndConsume(ctx context.Context) {
	go func() {
		if err := consumer.listenAndConsume(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("listening message deleted events: %v", err)
		}
	}()
}

func (consumer *MessageDeletedConsumer) listenAndConsume(ctx context.Context) error {
	pubsub := consumer.client.Subscribe(ctx, messageDeletedChannel)
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

			var event domain.MessageDeletedEvent
			if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
				log.Printf("unmarshalling message deleted event: %v", err)
				continue
			}

			for _, handler := range consumer.handlers {
				if err := handler(ctx, event); err != nil {
					log.Printf("handling message deleted event %q: %v", event.MessageUuid, err)
					continue
				}
			}
		}
	}
}

// Ensure MessageDeletedConsumer implements the message-deleted consumer contract.
var _ domain.MessageDeletedConsumer = (*MessageDeletedConsumer)(nil)
