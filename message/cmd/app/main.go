package main

import (
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
	"github.com/EugeneNail/vox/message/internal/application/usecases/open_chat"
	"github.com/EugeneNail/vox/message/internal/infrastructure/config"
	"github.com/EugeneNail/vox/message/internal/infrastructure/postgres"
	redis_infrastructure "github.com/EugeneNail/vox/message/internal/infrastructure/redis"
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

	// --- Section: Event delivery ---
	messageCreatedPublisher := redis_infrastructure.NewMessageCreatedPublisher(redisClient, configuration.Streams.MessageCreatedMaxLen)
	messageEditedPublisher := redis_infrastructure.NewMessageEditedPublisher(redisClient, configuration.Streams.MessageEditedMaxLen)
	messageDeletedPublisher := redis_infrastructure.NewMessageDeletedPublisher(redisClient, configuration.Streams.MessageDeletedMaxLen)
	userOpenedChatPublisher := redis_infrastructure.NewUserOpenedChatPublisher(redisClient)

	// --- Section: Application use-cases ---
	authorizeChatUpdatesHandler := authorize_chat_updates.NewHandler(chatRepository, chatMemberRepository)
	createChatHandler := create_chat.NewHandler(chatRepository)
	createMessageHandler := create_message.NewHandler(messageRepository, chatRepository, chatMemberRepository, messageCreatedPublisher)
	deleteMessageHandler := delete_message.NewHandler(messageRepository, messageDeletedPublisher)
	editMessageHandler := edit_message.NewHandler(messageRepository, messageEditedPublisher)
	openChatHandler := open_chat.NewHandler(authorizeChatUpdatesHandler, userOpenedChatPublisher)
	listChatMessagesHandler := list_chat_messages.NewHandler(messageRepository, chatRepository, chatMemberRepository)
	listChatsHandler := list_chats.NewHandler(chatRepository, chatMemberRepository)

	// --- Section: HTTP transport ---
	createChatHttpHandler := transport_http.NewCreateChatHandler(createChatHandler)
	createMessageHttpHandler := transport_http.NewCreateMessageHandler(createMessageHandler)
	deleteMessageHttpHandler := transport_http.NewDeleteMessageHandler(deleteMessageHandler)
	editMessageHttpHandler := transport_http.NewEditMessageHandler(editMessageHandler)
	openChatHttpHandler := transport_http.NewOpenChatHandler(openChatHandler)
	listChatMessagesHttpHandler := transport_http.NewListChatMessagesHandler(listChatMessagesHandler)
	listChatsHttpHandler := transport_http.NewListChatsHandler(listChatsHandler)

	// --- Section: HTTP routes ---
	webServer := http.NewServeMux()
	webServer.HandleFunc("POST   /api/v1/message/chats", middleware.RequireAuthenticatedUser(middleware.RejectLargeRequest(2048, middleware.WriteJsonResponse(createChatHttpHandler.Handle))))
	webServer.HandleFunc("GET    /api/v1/message/chats", middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(listChatsHttpHandler.Handle)))
	webServer.HandleFunc("POST   /api/v1/message/chats/{chatUuid}/open", middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(openChatHttpHandler.Handle)))
	webServer.HandleFunc("POST   /api/v1/message/chats/{chatUuid}/messages", middleware.RequireAuthenticatedUser(middleware.RejectLargeRequest(4096, middleware.WriteJsonResponse(createMessageHttpHandler.Handle))))
	webServer.HandleFunc("GET    /api/v1/message/chats/{chatUuid}/messages", middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(listChatMessagesHttpHandler.Handle)))
	webServer.HandleFunc("PUT    /api/v1/message/messages/{messageUuid}", middleware.RequireAuthenticatedUser(middleware.RejectLargeRequest(4096, middleware.WriteJsonResponse(editMessageHttpHandler.Handle))))
	webServer.HandleFunc("DELETE /api/v1/message/messages/{messageUuid}", middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(deleteMessageHttpHandler.Handle)))

	// --- Section: HTTP server ---
	address := fmt.Sprintf("0.0.0.0:%d", configuration.App.Port)

	log.Printf("message service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}
