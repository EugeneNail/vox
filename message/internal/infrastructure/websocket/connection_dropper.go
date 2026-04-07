package websocket

import "github.com/google/uuid"

// ConnectionDropper removes a websocket connection and every runtime state attached to it.
type ConnectionDropper struct {
	connectionHub        *ConnectionHub
	subscriptionRegistry *ChatSubscriptionRegistry
}

// NewConnectionDropper constructs a websocket connection dropper.
func NewConnectionDropper(connectionHub *ConnectionHub, subscriptionRegistry *ChatSubscriptionRegistry) *ConnectionDropper {
	return &ConnectionDropper{
		connectionHub:        connectionHub,
		subscriptionRegistry: subscriptionRegistry,
	}
}

// Drop removes the websocket connection and every runtime state attached to it.
func (dropper *ConnectionDropper) Drop(connectionUuid uuid.UUID) {
	connection := dropper.connectionHub.FindByUuid(connectionUuid)

	dropper.subscriptionRegistry.Unsubscribe(connectionUuid)
	dropper.connectionHub.Unregister(connectionUuid)

	if connection != nil {
		_ = connection.Close()
	}
}
