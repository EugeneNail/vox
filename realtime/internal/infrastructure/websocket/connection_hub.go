package websocket

import (
	"sync"

	"github.com/google/uuid"
	gorillawebsocket "github.com/gorilla/websocket"
)

// ConnectionHub stores active websocket connections.
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

// Register stores the websocket connection for future realtime delivery.
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

// FindByUserUuid finds all active websocket connections for the user.
func (hub *ConnectionHub) FindByUserUuid(userUuid uuid.UUID) []*Connection {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()

	connections := make([]*Connection, 0)
	for _, connection := range hub.connections {
		if connection.UserUuid() == userUuid {
			connections = append(connections, connection)
		}
	}

	return connections
}
