package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/auth/internal/domain/events"
	redisclient "github.com/redis/go-redis/v9"
)

const userCreatedStream = "user.created"

// UserCreatedPublisher publishes user-created events through Redis Streams.
type UserCreatedPublisher struct {
	client *redisclient.Client
}

// NewUserCreatedPublisher constructs a Redis-backed user-created publisher.
func NewUserCreatedPublisher(client *redisclient.Client) *UserCreatedPublisher {
	return &UserCreatedPublisher{
		client: client,
	}
}

// Publish publishes a user-created event.
func (publisher *UserCreatedPublisher) Publish(ctx context.Context, event events.UserCreated) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling user created event for user %q: %w", event.UserUuid, err)
	}

	if err := publisher.client.XAdd(ctx, &redisclient.XAddArgs{
		Stream: userCreatedStream,
		Values: map[string]any{
			"payload": string(payload),
		},
	}).Err(); err != nil {
		return fmt.Errorf("publishing user created event for user %q: %w", event.UserUuid, err)
	}

	return nil
}

var _ events.UserCreatedPublisher = (*UserCreatedPublisher)(nil)
