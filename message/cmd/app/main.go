package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/EugeneNail/vox/lib-common/http/middleware"
	"github.com/EugeneNail/vox/message/internal/application/usecases/authorize_direct_chat_updates"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_direct_chat"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_message"
	"github.com/EugeneNail/vox/message/internal/application/usecases/delete_message"
	"github.com/EugeneNail/vox/message/internal/application/usecases/edit_message"
	"github.com/EugeneNail/vox/message/internal/application/usecases/list_chat_messages"
	"github.com/EugeneNail/vox/message/internal/application/usecases/list_direct_chats"
	"github.com/EugeneNail/vox/message/internal/infrastructure/config"
	message_middleware "github.com/EugeneNail/vox/message/internal/infrastructure/http/middleware"
	"github.com/EugeneNail/vox/message/internal/infrastructure/postgres"
	redis_infrastructure "github.com/EugeneNail/vox/message/internal/infrastructure/redis"
	websocket_infrastructure "github.com/EugeneNail/vox/message/internal/infrastructure/websocket"
	transport_http "github.com/EugeneNail/vox/message/internal/transport/http"
)

func main() {
	// --- Section: Runtime configuration ---
	configuration, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	// --- Section: External clients ---
	database, err := postgres.NewDatabase(configuration.Postgres)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	redisClient, err := redis_infrastructure.NewClient(configuration.Redis)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()

	// --- Section: Repositories ---
	messageRepository := postgres.NewMessageRepository(database)
	directChatRepository := postgres.NewDirectChatRepository(database)

	// --- Section: WebSocket runtime ---
	connectionHub := websocket_infrastructure.NewConnectionHub()
	chatSubscriptionRegistry := websocket_infrastructure.NewChatSubscriptionRegistry()
	connectionDropper := websocket_infrastructure.NewConnectionDropper(connectionHub, chatSubscriptionRegistry)

	// --- Section: Event delivery ---
	messageCreatedPublisher := redis_infrastructure.NewMessageCreatedPublisher(redisClient)
	messageEditedPublisher := redis_infrastructure.NewMessageEditedPublisher(redisClient)
	messageDeletedPublisher := redis_infrastructure.NewMessageDeletedPublisher(redisClient)
	addMessageWebSocketSender := websocket_infrastructure.NewAddMessageWebSocketSender(connectionHub, chatSubscriptionRegistry, connectionDropper)
	updateMessageWebSocketSender := websocket_infrastructure.NewUpdateMessageWebSocketSender(connectionHub, chatSubscriptionRegistry, connectionDropper)
	removeMessageWebSocketSender := websocket_infrastructure.NewRemoveMessageWebSocketSender(connectionHub, chatSubscriptionRegistry, connectionDropper)
	messageCreatedRedisConsumer := redis_infrastructure.NewMessageCreatedConsumer(redisClient, addMessageWebSocketSender.Send)
	messageEditedRedisConsumer := redis_infrastructure.NewMessageEditedConsumer(redisClient, updateMessageWebSocketSender.Send)
	messageDeletedRedisConsumer := redis_infrastructure.NewMessageDeletedConsumer(redisClient, removeMessageWebSocketSender.Send)

	// --- Section: Application use-cases ---
	authorizeDirectChatUpdatesHandler := authorize_direct_chat_updates.NewHandler(directChatRepository)
	createDirectChatHandler := create_direct_chat.NewHandler(directChatRepository)
	createMessageHandler := create_message.NewHandler(messageRepository, directChatRepository, messageCreatedPublisher)
	deleteMessageHandler := delete_message.NewHandler(messageRepository, directChatRepository, messageDeletedPublisher)
	editMessageHandler := edit_message.NewHandler(messageRepository, directChatRepository, messageEditedPublisher)
	listChatMessagesHandler := list_chat_messages.NewHandler(messageRepository, directChatRepository)
	listDirectChatsHandler := list_direct_chats.NewHandler(directChatRepository)

	// --- Section: HTTP transport ---
	httpHandler := transport_http.NewHandler(authorizeDirectChatUpdatesHandler, createDirectChatHandler, createMessageHandler, deleteMessageHandler, editMessageHandler, listChatMessagesHandler, listDirectChatsHandler, connectionHub, chatSubscriptionRegistry, connectionDropper)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageCreatedRedisConsumer.ListenAndConsume(ctx)
	messageEditedRedisConsumer.ListenAndConsume(ctx)
	messageDeletedRedisConsumer.ListenAndConsume(ctx)

	webServer := http.NewServeMux()
	webServer.HandleFunc("POST /api/v1/message/direct-chats", message_middleware.RequireAuthenticatedUser(middleware.RejectLargeRequest(2048, middleware.WriteJsonResponse(httpHandler.CreateDirectChat))))
	webServer.HandleFunc("GET /api/v1/message/direct-chats", message_middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(httpHandler.ListDirectChats)))
	webServer.HandleFunc("POST /api/v1/message/chats/{chatUuid}/messages", message_middleware.RequireAuthenticatedUser(middleware.RejectLargeRequest(4096, middleware.WriteJsonResponse(httpHandler.CreateMessage))))
	webServer.HandleFunc("GET /api/v1/message/direct-chats/{directChatUuid}/messages", message_middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(httpHandler.ListChatMessages)))
	webServer.HandleFunc("PUT /api/v1/message/messages/{messageUuid}", message_middleware.RequireAuthenticatedUser(middleware.RejectLargeRequest(4096, middleware.WriteJsonResponse(httpHandler.EditMessage))))
	webServer.HandleFunc("DELETE /api/v1/message/messages/{messageUuid}", message_middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(httpHandler.DeleteMessage)))
	webServer.HandleFunc("GET /api/v1/message/ws", httpHandler.OpenWebSocket)

	address := fmt.Sprintf("0.0.0.0:%d", configuration.App.Port)

	log.Printf("message service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}
