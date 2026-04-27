package http

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"

	"github.com/EugeneNail/vox/lib-common/authentication"
	websocket_infrastructure "github.com/EugeneNail/vox/realtime/internal/infrastructure/websocket"
	"github.com/gorilla/websocket"
)

type OpenWebSocketHandler struct {
	connectionHub     *websocket_infrastructure.ConnectionHub
	connectionDropper *websocket_infrastructure.ConnectionDropper
}

func NewOpenWebSocketHandler(connectionHub *websocket_infrastructure.ConnectionHub, connectionDropper *websocket_infrastructure.ConnectionDropper) *OpenWebSocketHandler {
	return &OpenWebSocketHandler{
		connectionHub:     connectionHub,
		connectionDropper: connectionDropper,
	}
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

	connection := handler.connectionHub.Register(socket, userUuid)
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
		if _, _, err := connection.ReadMessage(); err != nil {
			return
		}
	}
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
