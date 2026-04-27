package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/events"
)

// MessageEditedWebSocketSender sends message-edited commands to websocket connections selected by subscriptions.
type MessageEditedWebSocketSender struct {
	connectionHub        *ConnectionHub
	subscriptionRegistry *ChatSubscriptionRegistry
	connectionDropper    *ConnectionDropper
}

// NewMessageEditedWebSocketSender constructs a message-edited websocket sender.
func NewMessageEditedWebSocketSender(connectionHub *ConnectionHub, subscriptionRegistry *ChatSubscriptionRegistry, connectionDropper *ConnectionDropper) *MessageEditedWebSocketSender {
	return &MessageEditedWebSocketSender{
		connectionHub:        connectionHub,
		subscriptionRegistry: subscriptionRegistry,
		connectionDropper:    connectionDropper,
	}
}

// Send sends a message-edited command to connections subscribed to the message chat.
func (sender *MessageEditedWebSocketSender) Send(ctx context.Context, event events.MessageEdited) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	payload, err := json.Marshal(map[string]any{
		"type": "MessageEdited",
		"data": event,
	})
	if err != nil {
		return fmt.Errorf("marshalling message edited websocket command for message %q: %w", event.MessageUuid, err)
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
