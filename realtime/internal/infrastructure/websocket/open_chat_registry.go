package websocket

import (
	"sync"

	"github.com/google/uuid"
)

// OpenChatRegistry stores the current runtime open chat for each connection.
type OpenChatRegistry struct {
	mutex                          sync.RWMutex
	subscribedChatUuidByConnection map[uuid.UUID]uuid.UUID
	connectionUuidsByChatUuid      map[uuid.UUID]map[uuid.UUID]struct{}
}

// NewOpenChatRegistry constructs an open chat registry.
func NewOpenChatRegistry() *OpenChatRegistry {
	return &OpenChatRegistry{
		subscribedChatUuidByConnection: make(map[uuid.UUID]uuid.UUID),
		connectionUuidsByChatUuid:      make(map[uuid.UUID]map[uuid.UUID]struct{}),
	}
}

// OpenChat stores the current open chat for the connection.
func (registry *OpenChatRegistry) OpenChat(connectionUuid uuid.UUID, chatUuid uuid.UUID) {
	registry.CloseChat(connectionUuid)

	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	if registry.connectionUuidsByChatUuid[chatUuid] == nil {
		registry.connectionUuidsByChatUuid[chatUuid] = make(map[uuid.UUID]struct{})
	}
	registry.connectionUuidsByChatUuid[chatUuid][connectionUuid] = struct{}{}
	registry.subscribedChatUuidByConnection[connectionUuid] = chatUuid
}

// CloseChat removes the current open chat for the connection.
func (registry *OpenChatRegistry) CloseChat(connectionUuid uuid.UUID) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	chatUuid, exists := registry.subscribedChatUuidByConnection[connectionUuid]
	if !exists {
		return
	}

	if registry.connectionUuidsByChatUuid[chatUuid] != nil {
		delete(registry.connectionUuidsByChatUuid[chatUuid], connectionUuid)
		if len(registry.connectionUuidsByChatUuid[chatUuid]) == 0 {
			delete(registry.connectionUuidsByChatUuid, chatUuid)
		}
	}

	delete(registry.subscribedChatUuidByConnection, connectionUuid)
}

// FindConnectionUuidsByChatUuid finds all runtime connections with the chat currently open.
func (registry *OpenChatRegistry) FindConnectionUuidsByChatUuid(chatUuid uuid.UUID) []uuid.UUID {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	connectionUuids := make([]uuid.UUID, 0, len(registry.connectionUuidsByChatUuid[chatUuid]))
	for connectionUuid := range registry.connectionUuidsByChatUuid[chatUuid] {
		connectionUuids = append(connectionUuids, connectionUuid)
	}

	return connectionUuids
}
