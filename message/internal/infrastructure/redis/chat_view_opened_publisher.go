package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/events"
	redisclient "github.com/redis/go-redis/v9"
)

// ChatViewOpenedPublisher publishes chat-view-opened events through Redis Streams.
type ChatViewOpenedPublisher struct {
	client *redisclient.Client
}

// NewChatViewOpenedPublisher constructs a Redis-backed chat-view-opened publisher.
func NewChatViewOpenedPublisher(client *redisclient.Client) *ChatViewOpenedPublisher {
	return &ChatViewOpenedPublisher{
		client: client,
	}
}

// Publish publishes a chat-view-opened event.
func (publisher *ChatViewOpenedPublisher) Publish(ctx context.Context, event events.ChatViewOpened) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling chat view opened event for chat %q and user %q: %w", event.ChatUuid, event.UserUuid, err)
	}

	if err := publisher.client.XAdd(ctx, &redisclient.XAddArgs{
		Stream: events.ChatViewOpenedStream,
		Values: map[string]any{
			"payload": string(payload),
		},
	}).Err(); err != nil {
		return fmt.Errorf("publishing chat view opened event for chat %q and user %q: %w", event.ChatUuid, event.UserUuid, err)
	}

	return nil
}

var _ events.ChatViewOpenedPublisher = (*ChatViewOpenedPublisher)(nil)
