package redis

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/EugeneNail/vox/lib-common/events"
	"github.com/EugeneNail/vox/lib-common/redisstream"
	websocket_infrastructure "github.com/EugeneNail/vox/realtime/internal/infrastructure/websocket"
	redisclient "github.com/redis/go-redis/v9"
)

// MessageCreatedConsumer consumes message-created events through Redis Streams.
type MessageCreatedConsumer struct {
	client       *redisclient.Client
	consumerName string
	sender       *websocket_infrastructure.MessageCreatedSender
}

// NewMessageCreatedConsumer constructs a Redis-backed message-created consumer.
func NewMessageCreatedConsumer(client *redisclient.Client, sender *websocket_infrastructure.MessageCreatedSender) *MessageCreatedConsumer {
	return &MessageCreatedConsumer{
		client:       client,
		consumerName: redisstream.BuildConsumerName("realtime-service"),
		sender:       sender,
	}
}

// ListenAndConsume starts message-created consumption in a goroutine and logs unexpected errors.
func (consumer *MessageCreatedConsumer) ListenAndConsume(ctx context.Context) {
	go func() {
		err := redisstream.ListenAndConsume(ctx, consumer.client, events.MessageCreatedStream, realtimeEventsConsumerGroup, consumer.consumerName, consumer.handlePayload)
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

	if err := consumer.sender.Send(ctx, event); err != nil {
		log.Printf("sending message created websocket event for message %q: %v", event.MessageUuid, err)
	}

	return true
}
