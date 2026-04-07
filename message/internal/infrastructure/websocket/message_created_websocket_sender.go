package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
)

// MessageCreatedWebSocketSender sends message-created events to websocket connections selected by subscriptions.
type MessageCreatedWebSocketSender struct {
	connectionHub        *ConnectionHub
	subscriptionRegistry *ChatSubscriptionRegistry
	connectionDropper    *ConnectionDropper
}

// NewMessageCreatedWebSocketSender constructs a message-created websocket sender.
func NewMessageCreatedWebSocketSender(connectionHub *ConnectionHub, subscriptionRegistry *ChatSubscriptionRegistry, connectionDropper *ConnectionDropper) *MessageCreatedWebSocketSender {
	return &MessageCreatedWebSocketSender{
		connectionHub:        connectionHub,
		subscriptionRegistry: subscriptionRegistry,
		connectionDropper:    connectionDropper,
	}
}

// Send sends a message-created event to connections subscribed to the message chat.
func (sender *MessageCreatedWebSocketSender) Send(ctx context.Context, event domain.MessageCreatedEvent) error {
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

	connectionUuids := sender.subscriptionRegistry.FindConnectionUuidsByChatUuid(event.ChatUuid)
	for _, connectionUuid := range connectionUuids {
		connection := sender.connectionHub.FindByUuid(connectionUuid)
		if connection == nil {
			sender.connectionDropper.Drop(connectionUuid)
			continue
		}

		if err := connection.WriteText(payload); err != nil {
			sender.connectionDropper.Drop(connection.Uuid())
			continue
		}
	}

	return nil
}
