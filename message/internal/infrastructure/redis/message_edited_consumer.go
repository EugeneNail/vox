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

type MessageEditedHandler func(context.Context, events.MessageEdited) error

// MessageEditedConsumer consumes message-edited events through Redis Streams.
type MessageEditedConsumer struct {
	client       *redisclient.Client
	consumerName string
	handlers     []MessageEditedHandler
}

// NewMessageEditedConsumer constructs a Redis-backed message-edited consumer.
func NewMessageEditedConsumer(client *redisclient.Client, handlers ...MessageEditedHandler) *MessageEditedConsumer {
	return &MessageEditedConsumer{
		client:       client,
		consumerName: redisstream.BuildConsumerName("message-service"),
		handlers:     handlers,
	}
}

// ListenAndConsume starts message-edited consumption in a goroutine and logs unexpected errors.
func (consumer *MessageEditedConsumer) ListenAndConsume(ctx context.Context) {
	go func() {
		err := redisstream.ListenAndConsume(ctx, consumer.client, events.MessageEditedStream, messageEventsConsumerGroup, consumer.consumerName, consumer.handlePayload)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("listening message edited events: %v", err)
		}
	}()
}

func (consumer *MessageEditedConsumer) handlePayload(ctx context.Context, payload string) bool {
	var event events.MessageEdited
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Printf("unmarshalling message edited event: %v", err)
		return true
	}

	for _, handler := range consumer.handlers {
		if err := handler(ctx, event); err != nil {
			log.Printf("handling message edited event %q: %v", event.MessageUuid, err)
			return true
		}
	}

	return true
}
