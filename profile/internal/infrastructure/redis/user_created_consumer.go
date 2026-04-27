package redis

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/EugeneNail/vox/lib-common/redisstream"
	"github.com/EugeneNail/vox/profile/internal/domain/events"
	redisclient "github.com/redis/go-redis/v9"
)

const (
	userCreatedStream       = "user.created"
	userEventsConsumerGroup = "profile-service"
)

type UserCreatedHandler func(context.Context, events.UserCreated) error

// UserCreatedConsumer consumes user-created events through Redis Streams.
type UserCreatedConsumer struct {
	client       *redisclient.Client
	consumerName string
	handlers     []UserCreatedHandler
}

// NewUserCreatedConsumer constructs a Redis-backed user-created consumer.
func NewUserCreatedConsumer(client *redisclient.Client, handlers ...UserCreatedHandler) *UserCreatedConsumer {
	return &UserCreatedConsumer{
		client:       client,
		consumerName: redisstream.BuildConsumerName("profile-service"),
		handlers:     handlers,
	}
}

// ListenAndConsume starts user-created consumption in a goroutine and logs unexpected errors.
func (consumer *UserCreatedConsumer) ListenAndConsume(ctx context.Context) {
	go func() {
		err := redisstream.ListenAndConsume(ctx, consumer.client, userCreatedStream, userEventsConsumerGroup, consumer.consumerName, consumer.handleMessage)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("listening user created events: %v", err)
		}
	}()
}

func (consumer *UserCreatedConsumer) handleMessage(ctx context.Context, payload string) bool {
	var event events.UserCreated
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Printf("unmarshalling user created event: %v", err)
		return true
	}

	for _, handler := range consumer.handlers {
		if err := handler(ctx, event); err != nil {
			log.Printf("handling user created event for user %q: %v", event.UserUuid, err)
			return true
		}
	}

	return true
}
