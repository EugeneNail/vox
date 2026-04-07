package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
)

// MessageRealtimeDispatcher sends message events to websocket connections selected by subscriptions.
type MessageRealtimeDispatcher struct {
	connectionHub        *ConnectionHub
	subscriptionRegistry *ChatSubscriptionRegistry
}

// NewMessageRealtimeDispatcher constructs a message realtime dispatcher.
func NewMessageRealtimeDispatcher(connectionHub *ConnectionHub, subscriptionRegistry *ChatSubscriptionRegistry) *MessageRealtimeDispatcher {
	return &MessageRealtimeDispatcher{
		connectionHub:        connectionHub,
		subscriptionRegistry: subscriptionRegistry,
	}
}

// DispatchMessageCreated sends a message-created event to connections subscribed to the message chat.
func (dispatcher *MessageRealtimeDispatcher) DispatchMessageCreated(ctx context.Context, event domain.MessageCreatedEvent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	payload, err := json.Marshal(map[string]any{
		"type": "message.created",
		"data": event,
	})
	if err != nil {
		return fmt.Errorf("marshalling message created websocket event %q: %w", event.MessageUuid, err)
	}

	connectionUuids := dispatcher.subscriptionRegistry.FindConnectionUuidsByChatUuid(event.ChatUuid)
	for _, connectionUuid := range connectionUuids {
		connection := dispatcher.connectionHub.FindByUuid(connectionUuid)
		if connection == nil {
			dispatcher.subscriptionRegistry.Unsubscribe(connectionUuid)
			continue
		}

		if err := connection.WriteText(payload); err != nil {
			dispatcher.connectionHub.Unregister(connectionUuid)
			dispatcher.subscriptionRegistry.Unsubscribe(connectionUuid)
			_ = connection.Close()
			continue
		}
	}

	return nil
}
