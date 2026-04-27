package websocket

import (
	"sync"

	"github.com/google/uuid"
)

// ChatSubscriptionRegistry stores the current runtime chat subscription for each connection.
type ChatSubscriptionRegistry struct {
	mutex                          sync.RWMutex
	subscribedChatUuidByConnection map[uuid.UUID]uuid.UUID
	connectionUuidsByChatUuid      map[uuid.UUID]map[uuid.UUID]struct{}
}

// NewChatSubscriptionRegistry constructs a chat subscription registry.
func NewChatSubscriptionRegistry() *ChatSubscriptionRegistry {
	return &ChatSubscriptionRegistry{
		subscribedChatUuidByConnection: make(map[uuid.UUID]uuid.UUID),
		connectionUuidsByChatUuid:      make(map[uuid.UUID]map[uuid.UUID]struct{}),
	}
}

// Subscribe stores the current chat subscription for the connection.
func (registry *ChatSubscriptionRegistry) Subscribe(connectionUuid uuid.UUID, chatUuid uuid.UUID) {
	registry.Unsubscribe(connectionUuid)

	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	if registry.connectionUuidsByChatUuid[chatUuid] == nil {
		registry.connectionUuidsByChatUuid[chatUuid] = make(map[uuid.UUID]struct{})
	}
	registry.connectionUuidsByChatUuid[chatUuid][connectionUuid] = struct{}{}
	registry.subscribedChatUuidByConnection[connectionUuid] = chatUuid
}

// Unsubscribe removes the current chat subscription for the connection.
func (registry *ChatSubscriptionRegistry) Unsubscribe(connectionUuid uuid.UUID) {
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

// FindConnectionUuidsByChatUuid finds all runtime connections subscribed to the chat.
func (registry *ChatSubscriptionRegistry) FindConnectionUuidsByChatUuid(chatUuid uuid.UUID) []uuid.UUID {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	connectionUuids := make([]uuid.UUID, 0, len(registry.connectionUuidsByChatUuid[chatUuid]))
	for connectionUuid := range registry.connectionUuidsByChatUuid[chatUuid] {
		connectionUuids = append(connectionUuids, connectionUuid)
	}

	return connectionUuids
}
