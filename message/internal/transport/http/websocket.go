package http

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	message_middleware "github.com/EugeneNail/vox/message/internal/infrastructure/http/middleware"
	"github.com/gorilla/websocket"
)

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

	connection, err := updatesWebSocketUpgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Printf("upgrading message websocket for user '%s': %v", userUuid, err)
		return
	}
	defer connection.Close()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-request.Context().Done():
			return
		case <-ticker.C:
			if err := connection.WriteMessage(websocket.TextMessage, []byte("pong")); err != nil {
				log.Printf("writing message websocket pong for user '%s': %v", userUuid, err)
				return
			}
		}
	}
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
