package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/events"
)

// MessageCreatedSender sends message-created websocket events to connections with the chat currently open.
type MessageCreatedSender struct {
	connectionHub     *ConnectionHub
	openChatRegistry  *OpenChatRegistry
	connectionDropper *ConnectionDropper
}

// NewMessageCreatedSender constructs a message-created websocket sender.
func NewMessageCreatedSender(connectionHub *ConnectionHub, openChatRegistry *OpenChatRegistry, connectionDropper *ConnectionDropper) *MessageCreatedSender {
	return &MessageCreatedSender{
		connectionHub:     connectionHub,
		openChatRegistry:  openChatRegistry,
		connectionDropper: connectionDropper,
	}
}

// Send sends a message-created event to connections with the chat currently open.
func (sender *MessageCreatedSender) Send(ctx context.Context, event events.MessageCreated) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	payload, err := json.Marshal(map[string]any{
		"type": "MessageCreated",
		"data": event,
	})
	if err != nil {
		return fmt.Errorf("marshalling message created websocket event for message %q: %w", event.MessageUuid, err)
	}

	connectionUuids := sender.openChatRegistry.FindConnectionUuidsByChatUuid(event.ChatUuid)
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
