package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/events"
)

// MessageCreatedSender sends message-created websocket events to all active websocket connections.
type MessageCreatedSender struct {
	connectionHub     *ConnectionHub
	connectionDropper *ConnectionDropper
}

// NewMessageCreatedSender constructs a message-created websocket sender.
func NewMessageCreatedSender(connectionHub *ConnectionHub, connectionDropper *ConnectionDropper) *MessageCreatedSender {
	return &MessageCreatedSender{
		connectionHub:     connectionHub,
		connectionDropper: connectionDropper,
	}
}

// Send sends a message-created event to all current websocket connections.
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

	for _, connection := range sender.connectionHub.FindAll() {
		if connection == nil {
			continue
		}

		if err := connection.WriteText(payload); err != nil {
			sender.connectionDropper.Drop(connection.Uuid())
			continue
		}
	}

	return nil
}
