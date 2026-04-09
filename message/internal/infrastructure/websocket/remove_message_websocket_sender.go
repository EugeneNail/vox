package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain/events"
)

// RemoveMessageWebSocketSender sends remove-message commands to websocket connections selected by subscriptions.
type RemoveMessageWebSocketSender struct {
	connectionHub        *ConnectionHub
	subscriptionRegistry *ChatSubscriptionRegistry
	connectionDropper    *ConnectionDropper
}

// NewRemoveMessageWebSocketSender constructs a remove-message websocket sender.
func NewRemoveMessageWebSocketSender(connectionHub *ConnectionHub, subscriptionRegistry *ChatSubscriptionRegistry, connectionDropper *ConnectionDropper) *RemoveMessageWebSocketSender {
	return &RemoveMessageWebSocketSender{
		connectionHub:        connectionHub,
		subscriptionRegistry: subscriptionRegistry,
		connectionDropper:    connectionDropper,
	}
}

// Send sends a remove-message command to connections subscribed to the message chat.
func (sender *RemoveMessageWebSocketSender) Send(ctx context.Context, event events.MessageDeleted) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	payload, err := json.Marshal(map[string]any{
		"type": "chat.remove_message",
		"data": event,
	})
	if err != nil {
		return fmt.Errorf("marshalling remove message websocket command for message %q: %w", event.MessageUuid, err)
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
