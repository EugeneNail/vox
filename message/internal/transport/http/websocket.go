package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/EugeneNail/vox/message/internal/application/usecases/authorize_direct_chat_updates"
	message_middleware "github.com/EugeneNail/vox/message/internal/infrastructure/http/middleware"
	websocket_infrastructure "github.com/EugeneNail/vox/message/internal/infrastructure/websocket"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type updatesWebSocketCommand struct {
	Type     string `json:"type"`
	ChatUuid string `json:"chatUuid"`
}

var updatesWebSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     isAllowedWebSocketOrigin,
}

// UpdatesWebSocket streams realtime message updates to the current browser tab.
func (handler *Handler) UpdatesWebSocket(writer http.ResponseWriter, request *http.Request) {
	token := request.URL.Query().Get("token")
	userUuid, err := message_middleware.UserUuidFromLoginToken(token)
	if err != nil {
		http.Error(writer, "invalid token", http.StatusUnauthorized)
		return
	}

	socket, err := updatesWebSocketUpgrader.Upgrade(writer, request, nil)
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

		if err := handler.handleUpdatesWebSocketCommand(request, connection, payload); err != nil {
			log.Printf("handling message websocket command for user '%s': %v", userUuid, err)
		}
	}
}

func (handler *Handler) handleUpdatesWebSocketCommand(request *http.Request, connection *websocket_infrastructure.Connection, payload []byte) error {
	var command updatesWebSocketCommand
	if err := json.Unmarshal(payload, &command); err != nil {
		return err
	}

	switch command.Type {
	case "chat.subscribe":
		chatUuid, err := uuid.Parse(command.ChatUuid)
		if err != nil {
			return err
		}

		if err := handler.authorizeDirectChatUpdatesHandler.Handle(request.Context(), authorize_direct_chat_updates.Query{
			DirectChatUuid: chatUuid,
			UserUuid:       connection.UserUuid(),
		}); err != nil {
			return fmt.Errorf("authorizing direct chat %q updates for user %q: %w", chatUuid, connection.UserUuid(), err)
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
