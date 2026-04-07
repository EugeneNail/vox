package websocket

import (
	"sync"

	"github.com/google/uuid"
	gorillawebsocket "github.com/gorilla/websocket"
)

// ConnectionHub stores active websocket connections without knowing subscription semantics.
type ConnectionHub struct {
	mutex       sync.RWMutex
	connections map[uuid.UUID]*Connection
}

// NewConnectionHub constructs a connection hub.
func NewConnectionHub() *ConnectionHub {
	return &ConnectionHub{
		connections: make(map[uuid.UUID]*Connection),
	}
}

// Register stores the websocket connection for future realtime event delivery.
func (hub *ConnectionHub) Register(socket *gorillawebsocket.Conn, userUuid uuid.UUID) *Connection {
	connection := &Connection{
		uuid:     uuid.New(),
		userUuid: userUuid,
		socket:   socket,
	}

	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	hub.connections[connection.uuid] = connection

	return connection
}

// Unregister removes the websocket connection from realtime delivery.
func (hub *ConnectionHub) Unregister(connectionUuid uuid.UUID) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock()

	delete(hub.connections, connectionUuid)
}

// FindByUuid finds an active websocket connection by runtime identifier.
func (hub *ConnectionHub) FindByUuid(connectionUuid uuid.UUID) *Connection {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()

	return hub.connections[connectionUuid]
}
