package redis

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"time"

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
		consumerName: buildConsumerName(),
		handlers:     handlers,
	}
}

// ListenAndConsume starts user-created consumption in a goroutine and logs unexpected errors.
func (consumer *UserCreatedConsumer) ListenAndConsume(ctx context.Context) {
	go func() {
		err := consumer.listenAndConsume(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("listening user created events: %v", err)
		}
	}()
}

func (consumer *UserCreatedConsumer) listenAndConsume(ctx context.Context) error {
	if err := ensureStreamGroup(ctx, consumer.client); err != nil {
		return err
	}

	for {
		streams, err := consumer.client.XReadGroup(ctx, &redisclient.XReadGroupArgs{
			Group:    userEventsConsumerGroup,
			Consumer: consumer.consumerName,
			Streams:  []string{userCreatedStream, ">"},
			Count:    10,
			Block:    5 * time.Second,
		}).Result()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}

			if errors.Is(err, redisclient.Nil) {
				continue
			}

			return err
		}

		for _, streamResult := range streams {
			for _, message := range streamResult.Messages {
				shouldAck := consumer.handleMessage(ctx, message)
				if !shouldAck {
					continue
				}

				if err := consumer.client.XAck(ctx, userCreatedStream, userEventsConsumerGroup, message.ID).Err(); err != nil {
					log.Printf("acknowledging user.created event %q: %v", message.ID, err)
				}
			}
		}
	}
}

func (consumer *UserCreatedConsumer) handleMessage(ctx context.Context, message redisclient.XMessage) bool {
	rawPayload, ok := message.Values["payload"]
	if !ok {
		log.Printf("user.created stream entry %q does not contain payload", message.ID)
		return true
	}

	payload, ok := rawPayload.(string)
	if !ok {
		log.Printf("user.created stream entry %q contains non-string payload", message.ID)
		return true
	}

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

func ensureStreamGroup(ctx context.Context, client *redisclient.Client) error {
	if err := client.XGroupCreateMkStream(ctx, userCreatedStream, userEventsConsumerGroup, "0").Err(); err != nil {
		if strings.Contains(err.Error(), "BUSYGROUP") {
			return nil
		}

		return err
	}

	return nil
}

func buildConsumerName() string {
	hostName, err := os.Hostname()
	if err != nil || hostName == "" {
		return "profile-service"
	}

	return hostName
}

var _ events.UserCreatedConsumer = (*UserCreatedConsumer)(nil)
