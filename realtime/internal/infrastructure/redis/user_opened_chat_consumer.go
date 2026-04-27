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

// UserOpenedChatConsumer consumes user-opened-chat events through Redis Streams.
type UserOpenedChatConsumer struct {
	client       *redisclient.Client
	consumerName string
	hub          *websocket_infrastructure.ConnectionHub
	registry     *websocket_infrastructure.ChatSubscriptionRegistry
}

// NewUserOpenedChatConsumer constructs a Redis-backed user-opened-chat consumer.
func NewUserOpenedChatConsumer(client *redisclient.Client, hub *websocket_infrastructure.ConnectionHub, registry *websocket_infrastructure.ChatSubscriptionRegistry) *UserOpenedChatConsumer {
	return &UserOpenedChatConsumer{
		client:       client,
		consumerName: redisstream.BuildConsumerName("realtime-service"),
		hub:          hub,
		registry:     registry,
	}
}

// ListenAndConsume starts user-opened-chat consumption in a goroutine and logs unexpected errors.
func (consumer *UserOpenedChatConsumer) ListenAndConsume(ctx context.Context) {
	go func() {
		err := redisstream.ListenAndConsume(ctx, consumer.client, events.UserOpenedChatStream, realtimeEventsConsumerGroup, consumer.consumerName, consumer.handlePayload)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("listening user opened chat events: %v", err)
		}
	}()
}

func (consumer *UserOpenedChatConsumer) handlePayload(ctx context.Context, payload string) bool {
	var event events.UserOpenedChat
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Printf("unmarshalling user opened chat event: %v", err)
		return true
	}

	connections := consumer.hub.FindByUserUuid(event.UserUuid)
	if len(connections) == 0 {
		return true
	}

	for _, connection := range connections {
		if connection == nil {
			continue
		}

		consumer.registry.Subscribe(connection.Uuid(), event.ChatUuid)
	}

	return true
}
