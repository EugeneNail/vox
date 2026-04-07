package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/EugeneNail/vox/lib-common/http/middleware"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_direct_chat"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_message"
	"github.com/EugeneNail/vox/message/internal/application/usecases/list_chat_messages"
	"github.com/EugeneNail/vox/message/internal/application/usecases/list_direct_chats"
	"github.com/EugeneNail/vox/message/internal/infrastructure/config"
	message_middleware "github.com/EugeneNail/vox/message/internal/infrastructure/http/middleware"
	"github.com/EugeneNail/vox/message/internal/infrastructure/postgres"
	transport_http "github.com/EugeneNail/vox/message/internal/transport/http"
)

func main() {
	configuration, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	database, err := postgres.NewDatabase(configuration.Postgres)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	messageRepository := postgres.NewMessageRepository(database)
	directChatRepository := postgres.NewDirectChatRepository(database)
	createDirectChatHandler := create_direct_chat.NewHandler(directChatRepository)
	createMessageHandler := create_message.NewHandler(messageRepository, directChatRepository)
	listChatMessagesHandler := list_chat_messages.NewHandler(messageRepository, directChatRepository)
	listDirectChatsHandler := list_direct_chats.NewHandler(directChatRepository)
	httpHandler := transport_http.NewHandler(createDirectChatHandler, createMessageHandler, listChatMessagesHandler, listDirectChatsHandler)

	webServer := http.NewServeMux()
	webServer.HandleFunc(
		"GET /api/v1/message/ping",
		middleware.WriteJsonResponse(httpHandler.Ping),
	)
	webServer.HandleFunc(
		"POST /api/v1/message/direct-chats",
		message_middleware.RequireAuthenticatedUser(
			middleware.RejectLargeRequest(2048, middleware.WriteJsonResponse(httpHandler.CreateDirectChat)),
		),
	)
	webServer.HandleFunc(
		"GET /api/v1/message/direct-chats",
		message_middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(httpHandler.ListDirectChats)),
	)
	webServer.HandleFunc(
		"POST /api/v1/message/chats/{chatUuid}/messages",
		message_middleware.RequireAuthenticatedUser(
			middleware.RejectLargeRequest(4096, middleware.WriteJsonResponse(httpHandler.CreateMessage)),
		),
	)
	webServer.HandleFunc(
		"GET /api/v1/message/direct-chats/{directChatUuid}/messages",
		message_middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(httpHandler.ListChatMessages)),
	)
	webServer.HandleFunc(
		"GET /api/v1/message/ws",
		httpHandler.UpdatesWebSocket,
	)

	address := fmt.Sprintf("0.0.0.0:%d", configuration.App.Port)

	log.Printf("message service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}
