package websocket

import "github.com/google/uuid"

// ConnectionDropper removes a websocket connection and every runtime state attached to it.
type ConnectionDropper struct {
	connectionHub    *ConnectionHub
	openChatRegistry *OpenChatRegistry
}

// NewConnectionDropper constructs a websocket connection dropper.
func NewConnectionDropper(connectionHub *ConnectionHub, openChatRegistry *OpenChatRegistry) *ConnectionDropper {
	return &ConnectionDropper{
		connectionHub:    connectionHub,
		openChatRegistry: openChatRegistry,
	}
}

// Drop removes the websocket connection and every runtime state attached to it.
func (dropper *ConnectionDropper) Drop(connectionUuid uuid.UUID) {
	connection := dropper.connectionHub.FindByUuid(connectionUuid)

	dropper.openChatRegistry.CloseChat(connectionUuid)
	dropper.connectionHub.Unregister(connectionUuid)

	if connection != nil {
		_ = connection.Close()
	}
}
