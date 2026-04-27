package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/events"
	redisclient "github.com/redis/go-redis/v9"
)

// UserOpenedChatPublisher publishes user-opened-chat events through Redis Streams.
type UserOpenedChatPublisher struct {
	client *redisclient.Client
}

// NewUserOpenedChatPublisher constructs a Redis-backed user-opened-chat publisher.
func NewUserOpenedChatPublisher(client *redisclient.Client) *UserOpenedChatPublisher {
	return &UserOpenedChatPublisher{
		client: client,
	}
}

// Publish publishes a user-opened-chat event.
func (publisher *UserOpenedChatPublisher) Publish(ctx context.Context, event events.UserOpenedChat) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling user opened chat event for chat %q and user %q: %w", event.ChatUuid, event.UserUuid, err)
	}

	if err := publisher.client.XAdd(ctx, &redisclient.XAddArgs{
		Stream: events.UserOpenedChatStream,
		Values: map[string]any{
			"payload": string(payload),
		},
	}).Err(); err != nil {
		return fmt.Errorf("publishing user opened chat event for chat %q and user %q: %w", event.ChatUuid, event.UserUuid, err)
	}

	return nil
}

var _ events.UserOpenedChatPublisher = (*UserOpenedChatPublisher)(nil)
