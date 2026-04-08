package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/EugeneNail/vox/lib-common/http/middleware"
	"github.com/EugeneNail/vox/message/internal/application/usecases/authorize_chat_updates"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_chat"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_message"
	"github.com/EugeneNail/vox/message/internal/application/usecases/delete_message"
	"github.com/EugeneNail/vox/message/internal/application/usecases/edit_message"
	"github.com/EugeneNail/vox/message/internal/application/usecases/list_chat_messages"
	"github.com/EugeneNail/vox/message/internal/application/usecases/list_chats"
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
	chatRepository := postgres.NewChatRepository(database)
	chatMemberRepository := postgres.NewChatMemberRepository(database)

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
	authorizeChatUpdatesHandler := authorize_chat_updates.NewHandler(chatRepository, chatMemberRepository)
	createChatHandler := create_chat.NewHandler(chatRepository)
	createMessageHandler := create_message.NewHandler(messageRepository, chatRepository, chatMemberRepository, messageCreatedPublisher)
	deleteMessageHandler := delete_message.NewHandler(messageRepository, chatRepository, chatMemberRepository, messageDeletedPublisher)
	editMessageHandler := edit_message.NewHandler(messageRepository, chatRepository, chatMemberRepository, messageEditedPublisher)
	listChatMessagesHandler := list_chat_messages.NewHandler(messageRepository, chatRepository, chatMemberRepository)
	listChatsHandler := list_chats.NewHandler(chatRepository, chatMemberRepository)

	// --- Section: HTTP transport ---
	createChatHttpHandler := transport_http.NewCreateChatHandler(createChatHandler)
	createMessageHttpHandler := transport_http.NewCreateMessageHandler(createMessageHandler)
	deleteMessageHttpHandler := transport_http.NewDeleteMessageHandler(deleteMessageHandler)
	editMessageHttpHandler := transport_http.NewEditMessageHandler(editMessageHandler)
	listChatMessagesHttpHandler := transport_http.NewListChatMessagesHandler(listChatMessagesHandler)
	listChatsHttpHandler := transport_http.NewListChatsHandler(listChatsHandler)
	openWebSocketHttpHandler := transport_http.NewOpenWebSocketHandler(authorizeChatUpdatesHandler, connectionHub, chatSubscriptionRegistry, connectionDropper)

	// --- Section: Event consumers ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageCreatedRedisConsumer.ListenAndConsume(ctx)
	messageEditedRedisConsumer.ListenAndConsume(ctx)
	messageDeletedRedisConsumer.ListenAndConsume(ctx)

	// --- Section: HTTP routes ---
	webServer := http.NewServeMux()
	webServer.HandleFunc("POST   /api/v1/message/chats", message_middleware.RequireAuthenticatedUser(middleware.RejectLargeRequest(2048, middleware.WriteJsonResponse(createChatHttpHandler.Handle))))
	webServer.HandleFunc("GET    /api/v1/message/chats", message_middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(listChatsHttpHandler.Handle)))
	webServer.HandleFunc("POST   /api/v1/message/chats/{chatUuid}/messages", message_middleware.RequireAuthenticatedUser(middleware.RejectLargeRequest(4096, middleware.WriteJsonResponse(createMessageHttpHandler.Handle))))
	webServer.HandleFunc("GET    /api/v1/message/chats/{chatUuid}/messages", message_middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(listChatMessagesHttpHandler.Handle)))
	webServer.HandleFunc("PUT    /api/v1/message/messages/{messageUuid}", message_middleware.RequireAuthenticatedUser(middleware.RejectLargeRequest(4096, middleware.WriteJsonResponse(editMessageHttpHandler.Handle))))
	webServer.HandleFunc("DELETE /api/v1/message/messages/{messageUuid}", message_middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(deleteMessageHttpHandler.Handle)))
	webServer.HandleFunc("GET    /api/v1/message/ws", openWebSocketHttpHandler.Handle)

	// --- Section: HTTP server ---
	address := fmt.Sprintf("0.0.0.0:%d", configuration.App.Port)

	log.Printf("message service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}
