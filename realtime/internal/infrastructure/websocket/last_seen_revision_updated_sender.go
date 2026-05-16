package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EugeneNail/vox/lib-common/events"
)

// LastSeenRevisionUpdatedSender sends last-seen-revision-updated websocket events to all active user connections.
type LastSeenRevisionUpdatedSender struct {
	connectionHub     *ConnectionHub
	connectionDropper *ConnectionDropper
}

// NewLastSeenRevisionUpdatedSender constructs a last-seen-revision-updated websocket sender.
func NewLastSeenRevisionUpdatedSender(connectionHub *ConnectionHub, connectionDropper *ConnectionDropper) *LastSeenRevisionUpdatedSender {
	return &LastSeenRevisionUpdatedSender{
		connectionHub:     connectionHub,
		connectionDropper: connectionDropper,
	}
}

// Send sends a last-seen-revision-updated event to all active websocket connections of the user.
func (sender *LastSeenRevisionUpdatedSender) Send(ctx context.Context, event events.LastSeenRevisionUpdated) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	payload, err := json.Marshal(map[string]any{
		"type": "LastSeenRevisionUpdated",
		"data": event,
	})
	if err != nil {
		return fmt.Errorf("marshalling last seen revision updated websocket event for chat %q and user %q: %w", event.ChatUuid, event.UserUuid, err)
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
