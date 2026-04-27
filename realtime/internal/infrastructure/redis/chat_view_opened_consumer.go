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

// ChatViewOpenedConsumer consumes chat-view-opened events through Redis Streams.
type ChatViewOpenedConsumer struct {
	client       *redisclient.Client
	consumerName string
	hub          *websocket_infrastructure.ConnectionHub
	registry     *websocket_infrastructure.ChatSubscriptionRegistry
}

// NewChatViewOpenedConsumer constructs a Redis-backed chat-view-opened consumer.
func NewChatViewOpenedConsumer(client *redisclient.Client, hub *websocket_infrastructure.ConnectionHub, registry *websocket_infrastructure.ChatSubscriptionRegistry) *ChatViewOpenedConsumer {
	return &ChatViewOpenedConsumer{
		client:       client,
		consumerName: redisstream.BuildConsumerName("realtime-service"),
		hub:          hub,
		registry:     registry,
	}
}

// ListenAndConsume starts chat-view-opened consumption in a goroutine and logs unexpected errors.
func (consumer *ChatViewOpenedConsumer) ListenAndConsume(ctx context.Context) {
	go func() {
		err := redisstream.ListenAndConsume(ctx, consumer.client, events.ChatViewOpenedStream, realtimeEventsConsumerGroup, consumer.consumerName, consumer.handlePayload)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("listening chat view opened events: %v", err)
		}
	}()
}

func (consumer *ChatViewOpenedConsumer) handlePayload(ctx context.Context, payload string) bool {
	var event events.ChatViewOpened
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Printf("unmarshalling chat view opened event: %v", err)
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
