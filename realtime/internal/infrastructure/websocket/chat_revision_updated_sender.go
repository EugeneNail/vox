package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/events"
)

// ChatRevisionUpdatedSender sends chat-revision-updated websocket events to all active user connections.
type ChatRevisionUpdatedSender struct {
	connectionHub     *ConnectionHub
	connectionDropper *ConnectionDropper
}

// NewChatRevisionUpdatedSender constructs a chat-revision-updated websocket sender.
func NewChatRevisionUpdatedSender(connectionHub *ConnectionHub, connectionDropper *ConnectionDropper) *ChatRevisionUpdatedSender {
	return &ChatRevisionUpdatedSender{
		connectionHub:     connectionHub,
		connectionDropper: connectionDropper,
	}
}

// Send sends a chat-revision-updated event to all active websocket connections of the user.
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

	connections := sender.connectionHub.FindByUserUuid(event.UserUuid)
	for _, connection := range connections {
		if err := connection.WriteText(payload); err != nil {
			sender.connectionDropper.Drop(connection.Uuid())
			continue
		}
	}

	return nil
}
