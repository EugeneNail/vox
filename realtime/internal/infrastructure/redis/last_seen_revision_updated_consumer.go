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

type LastSeenRevisionUpdatedHandler func(context.Context, events.LastSeenRevisionUpdated) error

// LastSeenRevisionUpdatedConsumer consumes last-seen-revision-updated events through Redis Streams.
type LastSeenRevisionUpdatedConsumer struct {
	client       *redisclient.Client
	consumerName string
	handlers     []LastSeenRevisionUpdatedHandler
}

// NewLastSeenRevisionUpdatedConsumer constructs a Redis-backed last-seen-revision-updated consumer.
func NewLastSeenRevisionUpdatedConsumer(client *redisclient.Client, handlers ...LastSeenRevisionUpdatedHandler) *LastSeenRevisionUpdatedConsumer {
	return &LastSeenRevisionUpdatedConsumer{
		client:       client,
		consumerName: redisstream.BuildConsumerName("realtime-service"),
		handlers:     handlers,
	}
}

// ListenAndConsume starts last-seen-revision-updated consumption in a goroutine and logs unexpected errors.
func (consumer *LastSeenRevisionUpdatedConsumer) ListenAndConsume(ctx context.Context) {
	go func() {
		err := redisstream.ListenAndConsume(ctx, consumer.client, events.LastSeenRevisionUpdatedStream, realtimeEventsConsumerGroup, consumer.consumerName, consumer.handlePayload)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("listening last seen revision updated events: %v", err)
		}
	}()
}

func (consumer *LastSeenRevisionUpdatedConsumer) handlePayload(ctx context.Context, payload string) bool {
	var event events.LastSeenRevisionUpdated
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Printf("unmarshalling last seen revision updated event: %v", err)
		return true
	}

	for _, handler := range consumer.handlers {
		if err := handler(ctx, event); err != nil {
			log.Printf("handling last seen revision updated event for chat %q and user %q: %v", event.ChatUuid, event.UserUuid, err)
			return true
		}
	}

	return true
}
