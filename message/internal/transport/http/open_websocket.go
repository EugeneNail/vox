package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/EugeneNail/vox/message/internal/application/usecases/authorize_chat_updates"
	message_middleware "github.com/EugeneNail/vox/message/internal/infrastructure/http/middleware"
	websocket_infrastructure "github.com/EugeneNail/vox/message/internal/infrastructure/websocket"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type openWebSocketCommand struct {
	Type     string `json:"type"`
	ChatUuid string `json:"chatUuid"`
}

type OpenWebSocketHandler struct {
	usecase              *authorize_chat_updates.Handler
	connectionHub        *websocket_infrastructure.ConnectionHub
	subscriptionRegistry *websocket_infrastructure.ChatSubscriptionRegistry
	connectionDropper    *websocket_infrastructure.ConnectionDropper
}

func NewOpenWebSocketHandler(usecase *authorize_chat_updates.Handler, connectionHub *websocket_infrastructure.ConnectionHub, subscriptionRegistry *websocket_infrastructure.ChatSubscriptionRegistry, connectionDropper *websocket_infrastructure.ConnectionDropper) *OpenWebSocketHandler {
	return &OpenWebSocketHandler{
		usecase:              usecase,
		connectionHub:        connectionHub,
		subscriptionRegistry: subscriptionRegistry,
		connectionDropper:    connectionDropper,
	}
}

var openWebSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     isAllowedWebSocketOrigin,
}

// OpenWebSocket streams realtime message updates to the current browser tab.
func (handler *OpenWebSocketHandler) Handle(writer http.ResponseWriter, request *http.Request) {
	token := request.URL.Query().Get("token")
	userUuid, err := message_middleware.UserUuidFromLoginToken(token)
	if err != nil {
		http.Error(writer, "invalid token", http.StatusUnauthorized)
		return
	}

	socket, err := openWebSocketUpgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Printf("upgrading message websocket for user '%s': %v", userUuid, err)
		return
	}

	connection := handler.connectionHub.Register(socket, userUuid)
	defer handler.connectionDropper.Drop(connection.Uuid())

	for {
		_, payload, err := connection.ReadMessage()
		if err != nil {
			log.Printf("reading message websocket for user '%s': %v", userUuid, err)
			return
		}

		if err := handler.handleOpenWebSocketCommand(request, connection, payload); err != nil {
			log.Printf("handling message websocket command for user '%s': %v", userUuid, err)
		}
	}
}

func (handler *OpenWebSocketHandler) handleOpenWebSocketCommand(request *http.Request, connection *websocket_infrastructure.Connection, payload []byte) error {
	var command openWebSocketCommand
	if err := json.Unmarshal(payload, &command); err != nil {
		return err
	}

	switch command.Type {
	case "chat.subscribe":
		chatUuid, err := uuid.Parse(command.ChatUuid)
		if err != nil {
			return err
		}

		if err := handler.usecase.Handle(request.Context(), authorize_chat_updates.Query{
			ChatUuid: chatUuid,
			UserUuid: connection.UserUuid(),
		}); err != nil {
			return fmt.Errorf("authorizing chat %q updates for user %q: %w", chatUuid, connection.UserUuid(), err)
		}

		handler.subscriptionRegistry.Subscribe(connection.Uuid(), chatUuid)
	case "chat.unsubscribe":
		handler.subscriptionRegistry.Unsubscribe(connection.Uuid())
	default:
		return nil
	}

	return nil
}

func isAllowedWebSocketOrigin(request *http.Request) bool {
	origin := request.Header.Get("Origin")
	if origin == "" {
		return true
	}

	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}

	requestHost := request.Host
	hostWithoutPort, _, err := net.SplitHostPort(requestHost)
	if err == nil {
		requestHost = hostWithoutPort
	}

	return originURL.Hostname() == requestHost
}
