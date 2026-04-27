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

type MessageDeletedHandler func(context.Context, events.MessageDeleted) error

// MessageDeletedConsumer consumes message-deleted events through Redis Streams.
type MessageDeletedConsumer struct {
	client       *redisclient.Client
	consumerName string
	handlers     []MessageDeletedHandler
}

// NewMessageDeletedConsumer constructs a Redis-backed message-deleted consumer.
func NewMessageDeletedConsumer(client *redisclient.Client, handlers ...MessageDeletedHandler) *MessageDeletedConsumer {
	return &MessageDeletedConsumer{
		client:       client,
		consumerName: redisstream.BuildConsumerName("message-service"),
		handlers:     handlers,
	}
}

// ListenAndConsume starts message-deleted consumption in a goroutine and logs unexpected errors.
func (consumer *MessageDeletedConsumer) ListenAndConsume(ctx context.Context) {
	go func() {
		err := redisstream.ListenAndConsume(ctx, consumer.client, messageDeletedStream, messageEventsConsumerGroup, consumer.consumerName, consumer.handlePayload)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("listening message deleted events: %v", err)
		}
	}()
}

func (consumer *MessageDeletedConsumer) handlePayload(ctx context.Context, payload string) bool {
	var event events.MessageDeleted
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Printf("unmarshalling message deleted event: %v", err)
		return true
	}

	for _, handler := range consumer.handlers {
		if err := handler(ctx, event); err != nil {
			log.Printf("handling message deleted event %q: %v", event.MessageUuid, err)
			return true
		}
	}

	return true
}
