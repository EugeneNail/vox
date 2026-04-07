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
	connectionDropper    *ConnectionDropper
}

// NewMessageRealtimeDispatcher constructs a message realtime dispatcher.
func NewMessageRealtimeDispatcher(connectionHub *ConnectionHub, subscriptionRegistry *ChatSubscriptionRegistry, connectionDropper *ConnectionDropper) *MessageRealtimeDispatcher {
	return &MessageRealtimeDispatcher{
		connectionHub:        connectionHub,
		subscriptionRegistry: subscriptionRegistry,
		connectionDropper:    connectionDropper,
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
			dispatcher.connectionDropper.Drop(connectionUuid)
			continue
		}

		if err := connection.WriteText(payload); err != nil {
			dispatcher.connectionDropper.Drop(connection.Uuid())
			continue
		}
	}

	return nil
}
