package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/events"
)

// ChatRevisionUpdatedSender sends chat-revision-updated websocket events to chat subscribers.
type ChatRevisionUpdatedSender struct {
	connectionHub        *ConnectionHub
	subscriptionRegistry *ChatSubscriptionRegistry
	connectionDropper    *ConnectionDropper
}

// NewChatRevisionUpdatedSender constructs a chat-revision-updated websocket sender.
func NewChatRevisionUpdatedSender(connectionHub *ConnectionHub, subscriptionRegistry *ChatSubscriptionRegistry, connectionDropper *ConnectionDropper) *ChatRevisionUpdatedSender {
	return &ChatRevisionUpdatedSender{
		connectionHub:        connectionHub,
		subscriptionRegistry: subscriptionRegistry,
		connectionDropper:    connectionDropper,
	}
}

// Send sends a chat-revision-updated event to connections subscribed to the chat.
func (sender *ChatRevisionUpdatedSender) Send(ctx context.Context, event events.ChatRevisionUpdated) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	payload, err := json.Marshal(map[string]any{
		"type": "ChatRevisionUpdated",
		"data": event,
	})
	if err != nil {
		return fmt.Errorf("marshalling chat revision updated websocket event for chat %q: %w", event.ChatUuid, err)
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
