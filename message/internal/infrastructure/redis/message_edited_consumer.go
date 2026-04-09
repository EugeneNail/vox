package redis

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/EugeneNail/vox/message/internal/domain/events"
	redisclient "github.com/redis/go-redis/v9"
)

type MessageEditedHandler func(context.Context, events.MessageEdited) error

// MessageEditedConsumer consumes message-edited events through Redis Streams.
type MessageEditedConsumer struct {
	client       *redisclient.Client
	consumerName string
	handlers     []MessageEditedHandler
}

// NewMessageEditedConsumer constructs a Redis-backed message-edited consumer.
func NewMessageEditedConsumer(client *redisclient.Client, handlers ...MessageEditedHandler) *MessageEditedConsumer {
	return &MessageEditedConsumer{
		client:       client,
		consumerName: buildConsumerName(),
		handlers:     handlers,
	}
}

// ListenAndConsume starts message-edited consumption in a goroutine and logs unexpected errors.
func (consumer *MessageEditedConsumer) ListenAndConsume(ctx context.Context) {
	go func() {
		err := listenAndConsumeStream(ctx, consumer.client, messageEditedStream, consumer.consumerName, consumer.handlePayload)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("listening message edited events: %v", err)
		}
	}()
}

func (consumer *MessageEditedConsumer) handlePayload(ctx context.Context, payload string) bool {
	var event events.MessageEdited
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Printf("unmarshalling message edited event: %v", err)
		return true
	}

	for _, handler := range consumer.handlers {
		if err := handler(ctx, event); err != nil {
			log.Printf("handling message edited event %q: %v", event.MessageUuid, err)
			return true
		}
	}

	return true
}

// Ensure MessageEditedConsumer implements the message-edited consumer contract.
var _ events.MessageEditedConsumer = (*MessageEditedConsumer)(nil)
