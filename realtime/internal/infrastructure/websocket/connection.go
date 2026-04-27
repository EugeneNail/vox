package websocket

import (
	"sync"

	"github.com/google/uuid"
	gorillawebsocket "github.com/gorilla/websocket"
)

// Connection wraps a websocket connection with runtime metadata and write locking.
type Connection struct {
	uuid     uuid.UUID
	userUuid uuid.UUID
	socket   *gorillawebsocket.Conn
	mutex    sync.Mutex
}

// Uuid returns the runtime identifier of the connection.
func (connection *Connection) Uuid() uuid.UUID {
	return connection.uuid
}

// UserUuid returns the authenticated user identifier attached to the connection.
func (connection *Connection) UserUuid() uuid.UUID {
	return connection.userUuid
}

// ReadMessage reads the next websocket message.
func (connection *Connection) ReadMessage() (int, []byte, error) {
	return connection.socket.ReadMessage()
}

// WriteText writes a text websocket message.
func (connection *Connection) WriteText(payload []byte) error {
	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	return connection.socket.WriteMessage(gorillawebsocket.TextMessage, payload)
}

// Close closes the underlying websocket connection.
func (connection *Connection) Close() error {
	return connection.socket.Close()
}
