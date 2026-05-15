package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/EugeneNail/vox/lib-common/authentication"
	message_infrastructure "github.com/EugeneNail/vox/realtime/internal/infrastructure/message"
	websocket_infrastructure "github.com/EugeneNail/vox/realtime/internal/infrastructure/websocket"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type OpenWebSocketHandler struct {
	connectionHub             *websocket_infrastructure.ConnectionHub
	connectionDropper         *websocket_infrastructure.ConnectionDropper
	subscriptionRegistry      *websocket_infrastructure.ChatSubscriptionRegistry
	authorizeChatAccessClient *message_infrastructure.AuthorizeChatAccessClient
}

func NewOpenWebSocketHandler(
	connectionHub *websocket_infrastructure.ConnectionHub,
	connectionDropper *websocket_infrastructure.ConnectionDropper,
	subscriptionRegistry *websocket_infrastructure.ChatSubscriptionRegistry,
	authorizeChatAccessClient *message_infrastructure.AuthorizeChatAccessClient,
) *OpenWebSocketHandler {
	return &OpenWebSocketHandler{
		connectionHub:             connectionHub,
		connectionDropper:         connectionDropper,
		subscriptionRegistry:      subscriptionRegistry,
		authorizeChatAccessClient: authorizeChatAccessClient,
	}
}

type websocketCommand struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type openChatCommandData struct {
	ChatUuid uuid.UUID `json:"chatUuid"`
}

var openWebSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     isAllowedWebSocketOrigin,
}

// Handle upgrades the request to a websocket connection and keeps it open.
func (handler *OpenWebSocketHandler) Handle(writer http.ResponseWriter, request *http.Request) {
	token := request.URL.Query().Get("token")
	userUuid, err := authentication.UserUuidFromLoginToken(token)
	if err != nil {
		http.Error(writer, "invalid token", http.StatusUnauthorized)
		return
	}

	socket, err := openWebSocketUpgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Printf("upgrading realtime websocket for user '%s': %v", userUuid, err)
		return
	}

	connection := handler.connectionHub.Register(socket, userUuid, token)
	log.Printf("realtime websocket connected: user=%s connection=%s", userUuid, connection.Uuid())
	defer func() {
		log.Printf("realtime websocket disconnected: user=%s connection=%s", userUuid, connection.Uuid())
		handler.connectionDropper.Drop(connection.Uuid())
	}()

	if err := connection.WriteText(mustMarshal(map[string]any{
		"type": "connected",
		"data": map[string]any{
			"connectionUuid": connection.Uuid(),
			"userUuid":       userUuid,
		},
	})); err != nil {
		log.Printf("sending realtime websocket ack for user '%s' connection '%s': %v", userUuid, connection.Uuid(), err)
		return
	}

	for {
		_, payload, err := connection.ReadMessage()
		if err != nil {
			return
		}

		if err := handler.handleMessage(request.Context(), connection, payload); err != nil {
			log.Printf("handling realtime websocket command for user '%s' connection '%s': %v", userUuid, connection.Uuid(), err)
		}
	}
}

func (handler *OpenWebSocketHandler) handleMessage(ctx context.Context, connection *websocket_infrastructure.Connection, payload []byte) error {
	var command websocketCommand
	if err := json.Unmarshal(payload, &command); err != nil {
		return fmt.Errorf("unmarshalling websocket command: %w", err)
	}

	switch command.Type {
	case "OpenChat":
		return handler.handleOpenChatCommand(ctx, connection, command.Data)
	case "UnsubscribeChat":
		handler.subscriptionRegistry.Unsubscribe(connection.Uuid())
		return nil
	default:
		return nil
	}
}

func (handler *OpenWebSocketHandler) handleOpenChatCommand(ctx context.Context, connection *websocket_infrastructure.Connection, payload json.RawMessage) error {
	var data openChatCommandData
	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("unmarshalling open chat websocket command: %w", err)
	}

	if data.ChatUuid == uuid.Nil {
		return errors.New("open chat websocket command requires chatUuid")
	}

	if err := handler.authorizeChatAccessClient.Authorize(ctx, data.ChatUuid, connection.LoginToken()); err != nil {
		return fmt.Errorf("authorizing websocket chat access for chat %q: %w", data.ChatUuid, err)
	}

	handler.subscriptionRegistry.Subscribe(connection.Uuid(), data.ChatUuid)
	return nil
}

func mustMarshal(value any) []byte {
	payload, err := json.Marshal(value)
	if err != nil {
		return []byte(`{"type":"connected"}`)
	}

	return payload
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
