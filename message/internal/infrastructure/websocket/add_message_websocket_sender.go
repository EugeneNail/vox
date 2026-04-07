package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/message/internal/domain"
)

// AddMessageWebSocketSender sends add-message commands to websocket connections selected by subscriptions.
type AddMessageWebSocketSender struct {
	connectionHub        *ConnectionHub
	subscriptionRegistry *ChatSubscriptionRegistry
	connectionDropper    *ConnectionDropper
}

// NewAddMessageWebSocketSender constructs an add-message websocket sender.
func NewAddMessageWebSocketSender(connectionHub *ConnectionHub, subscriptionRegistry *ChatSubscriptionRegistry, connectionDropper *ConnectionDropper) *AddMessageWebSocketSender {
	return &AddMessageWebSocketSender{
		connectionHub:        connectionHub,
		subscriptionRegistry: subscriptionRegistry,
		connectionDropper:    connectionDropper,
	}
}

// Send sends an add-message command to connections subscribed to the message chat.
func (sender *AddMessageWebSocketSender) Send(ctx context.Context, event domain.MessageCreatedEvent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	payload, err := json.Marshal(map[string]any{
		"type": "chat.add_message",
		"data": event,
	})
	if err != nil {
		return fmt.Errorf("marshalling add message websocket command for message %q: %w", event.MessageUuid, err)
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
