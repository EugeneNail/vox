package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/EugeneNail/vox/realtime/internal/infrastructure/config"
	message_infrastructure "github.com/EugeneNail/vox/realtime/internal/infrastructure/message"
	redis_infrastructure "github.com/EugeneNail/vox/realtime/internal/infrastructure/redis"
	websocket_infrastructure "github.com/EugeneNail/vox/realtime/internal/infrastructure/websocket"
	transport_http "github.com/EugeneNail/vox/realtime/internal/transport/http"
)

func main() {
	// --- Section: Runtime configuration ---
	configuration, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	// --- Section: External clients ---
	redisClient, err := redis_infrastructure.NewClient(configuration.Redis)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()

	// --- Section: WebSocket runtime ---
	connectionHub := websocket_infrastructure.NewConnectionHub()
	chatSubscriptionRegistry := websocket_infrastructure.NewChatSubscriptionRegistry()
	connectionDropper := websocket_infrastructure.NewConnectionDropper(connectionHub, chatSubscriptionRegistry)
	authorizeChatAccessClient := message_infrastructure.NewAuthorizeChatAccessClient(configuration.Message.BaseURL)

	// --- Section: HTTP transport ---
	openWebSocketHttpHandler := transport_http.NewOpenWebSocketHandler(
		connectionHub,
		connectionDropper,
		chatSubscriptionRegistry,
		authorizeChatAccessClient,
	)

	// --- Section: Event delivery ---
	messageCreatedSender := websocket_infrastructure.NewMessageCreatedSender(connectionHub, chatSubscriptionRegistry, connectionDropper)
	chatRevisionUpdatedSender := websocket_infrastructure.NewChatRevisionUpdatedSender(connectionHub, chatSubscriptionRegistry, connectionDropper)
	lastSeenRevisionUpdatedSender := websocket_infrastructure.NewLastSeenRevisionUpdatedSender(connectionHub, connectionDropper)
	messageEditedSender := websocket_infrastructure.NewMessageEditedWebSocketSender(connectionHub, chatSubscriptionRegistry, connectionDropper)
	messageDeletedSender := websocket_infrastructure.NewMessageDeletedWebSocketSender(connectionHub, chatSubscriptionRegistry, connectionDropper)
	messageCreatedRedisConsumer := redis_infrastructure.NewMessageCreatedConsumer(redisClient, messageCreatedSender)
	chatRevisionUpdatedRedisConsumer := redis_infrastructure.NewChatRevisionUpdatedConsumer(redisClient, chatRevisionUpdatedSender.Send)
	lastSeenRevisionUpdatedRedisConsumer := redis_infrastructure.NewLastSeenRevisionUpdatedConsumer(redisClient, lastSeenRevisionUpdatedSender.Send)
	messageEditedRedisConsumer := redis_infrastructure.NewMessageEditedConsumer(redisClient, messageEditedSender.Send)
	messageDeletedRedisConsumer := redis_infrastructure.NewMessageDeletedConsumer(redisClient, messageDeletedSender.Send)

	// --- Section: Event consumers ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageCreatedRedisConsumer.ListenAndConsume(ctx)
	chatRevisionUpdatedRedisConsumer.ListenAndConsume(ctx)
	lastSeenRevisionUpdatedRedisConsumer.ListenAndConsume(ctx)
	messageEditedRedisConsumer.ListenAndConsume(ctx)
	messageDeletedRedisConsumer.ListenAndConsume(ctx)

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
