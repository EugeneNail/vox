package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/EugeneNail/vox/realtime/internal/infrastructure/config"
	websocket_infrastructure "github.com/EugeneNail/vox/realtime/internal/infrastructure/websocket"
	transport_http "github.com/EugeneNail/vox/realtime/internal/transport/http"
)

func main() {
	// --- Section: Runtime configuration ---
	configuration, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	// --- Section: WebSocket runtime ---
	connectionHub := websocket_infrastructure.NewConnectionHub()
	connectionDropper := websocket_infrastructure.NewConnectionDropper(connectionHub)

	// --- Section: HTTP transport ---
	openWebSocketHttpHandler := transport_http.NewOpenWebSocketHandler(connectionHub, connectionDropper)

	// --- Section: HTTP routes ---
	webServer := http.NewServeMux()
	webServer.HandleFunc("GET    /api/v1/realtime/ws", openWebSocketHttpHandler.Handle)

	// --- Section: HTTP server ---
	address := fmt.Sprintf("0.0.0.0:%d", configuration.App.Port)

	log.Printf("realtime service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}
