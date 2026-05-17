package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/events"
)

// MessageDeletedWebSocketSender sends message-deleted commands to connections with the chat currently open.
type MessageDeletedWebSocketSender struct {
	connectionHub     *ConnectionHub
	openChatRegistry  *OpenChatRegistry
	connectionDropper *ConnectionDropper
}

// NewMessageDeletedWebSocketSender constructs a message-deleted websocket sender.
func NewMessageDeletedWebSocketSender(connectionHub *ConnectionHub, openChatRegistry *OpenChatRegistry, connectionDropper *ConnectionDropper) *MessageDeletedWebSocketSender {
	return &MessageDeletedWebSocketSender{
		connectionHub:     connectionHub,
		openChatRegistry:  openChatRegistry,
		connectionDropper: connectionDropper,
	}
}

// Send sends a message-deleted command to connections with the message chat currently open.
func (sender *MessageDeletedWebSocketSender) Send(ctx context.Context, event events.MessageDeleted) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	payload, err := json.Marshal(map[string]any{
		"type": "MessageDeleted",
		"data": event,
	})
	if err != nil {
		return fmt.Errorf("marshalling message deleted websocket command for message %q: %w", event.MessageUuid, err)
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
