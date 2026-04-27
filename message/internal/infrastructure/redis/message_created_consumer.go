package redis

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/EugeneNail/vox/lib-common/events"
	"github.com/EugeneNail/vox/lib-common/redisstream"
	redisclient "github.com/redis/go-redis/v9"
)

type MessageCreatedHandler func(context.Context, events.MessageCreated) error

// MessageCreatedConsumer consumes message-created events through Redis Streams.
type MessageCreatedConsumer struct {
	client       *redisclient.Client
	consumerName string
	handlers     []MessageCreatedHandler
}

// NewMessageCreatedConsumer constructs a Redis-backed message-created consumer.
func NewMessageCreatedConsumer(client *redisclient.Client, handlers ...MessageCreatedHandler) *MessageCreatedConsumer {
	return &MessageCreatedConsumer{
		client:       client,
		consumerName: redisstream.BuildConsumerName("message-service"),
		handlers:     handlers,
	}
}

// ListenAndConsume starts message-created consumption in a goroutine and logs unexpected errors.
func (consumer *MessageCreatedConsumer) ListenAndConsume(ctx context.Context) {
	go func() {
		err := redisstream.ListenAndConsume(ctx, consumer.client, events.MessageCreatedStream, messageEventsConsumerGroup, consumer.consumerName, consumer.handlePayload)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("listening message created events: %v", err)
		}
	}()
}

func (consumer *MessageCreatedConsumer) handlePayload(ctx context.Context, payload string) bool {
	var event events.MessageCreated
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Printf("unmarshalling message created event: %v", err)
		return true
	}

	for _, handler := range consumer.handlers {
		if err := handler(ctx, event); err != nil {
			log.Printf("handling message created event %q: %v", event.MessageUuid, err)
			return true
		}
	}

	return true
}
