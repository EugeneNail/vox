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

type ChatRevisionUpdatedHandler func(context.Context, events.ChatRevisionUpdated) error

// ChatRevisionUpdatedConsumer consumes chat-revision-updated events through Redis Streams.
type ChatRevisionUpdatedConsumer struct {
	client       *redisclient.Client
	consumerName string
	handlers     []ChatRevisionUpdatedHandler
}

// NewChatRevisionUpdatedConsumer constructs a Redis-backed chat-revision-updated consumer.
func NewChatRevisionUpdatedConsumer(client *redisclient.Client, handlers ...ChatRevisionUpdatedHandler) *ChatRevisionUpdatedConsumer {
	return &ChatRevisionUpdatedConsumer{
		client:       client,
		consumerName: redisstream.BuildConsumerName("realtime-service"),
		handlers:     handlers,
	}
}

// ListenAndConsume starts chat-revision-updated consumption in a goroutine and logs unexpected errors.
func (consumer *ChatRevisionUpdatedConsumer) ListenAndConsume(ctx context.Context) {
	go func() {
		err := redisstream.ListenAndConsume(ctx, consumer.client, events.ChatRevisionUpdatedStream, realtimeEventsConsumerGroup, consumer.consumerName, consumer.handlePayload)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("listening chat revision updated events: %v", err)
		}
	}()
}

func (consumer *ChatRevisionUpdatedConsumer) handlePayload(ctx context.Context, payload string) bool {
	var event events.ChatRevisionUpdated
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Printf("unmarshalling chat revision updated event: %v", err)
		return true
	}

	for _, handler := range consumer.handlers {
		if err := handler(ctx, event); err != nil {
			log.Printf("handling chat revision updated event for chat %q: %v", event.ChatUuid, err)
			return true
		}
	}

	return true
}
