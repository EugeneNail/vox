package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/EugeneNail/vox/lib-common/http/middleware"
	"github.com/EugeneNail/vox/message/internal/application/usecases/add_chat_members"
	"github.com/EugeneNail/vox/message/internal/application/usecases/authorize_chat_access"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_direct_chat"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_group_chat"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_message"
	"github.com/EugeneNail/vox/message/internal/application/usecases/delete_message"
	"github.com/EugeneNail/vox/message/internal/application/usecases/edit_message"
	"github.com/EugeneNail/vox/message/internal/application/usecases/kick_chat_member"
	"github.com/EugeneNail/vox/message/internal/application/usecases/list_chat_messages"
	"github.com/EugeneNail/vox/message/internal/application/usecases/list_chats"
	"github.com/EugeneNail/vox/message/internal/application/usecases/set_last_seen_revision"
	"github.com/EugeneNail/vox/message/internal/application/usecases/update_chat"
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
	lastSeenRevisionUpdatedPublisher := redis_infrastructure.NewLastSeenRevisionUpdatedPublisher(redisClient, configuration.Streams.LastSeenRevisionUpdatedMaxLen)

	// --- Section: Application use-cases ---
	authorizeChatAccessHandler := authorize_chat_access.NewHandler(chatRepository, chatMemberRepository)
	createDirectChatHandler := create_direct_chat.NewHandler(chatRepository)
	createGroupChatHandler := create_group_chat.NewHandler(chatRepository)
	addChatMembersHandler := add_chat_members.NewHandler(chatRepository, chatMemberRepository)
	kickChatMemberHandler := kick_chat_member.NewHandler(chatRepository, chatMemberRepository)
	createMessageHandler := create_message.NewHandler(
		messageRepository,
		chatRepository,
		chatMemberRepository,
		lastSeenRevisionUpdatedPublisher,
		messageCreatedPublisher,
	)
	deleteMessageHandler := delete_message.NewHandler(messageRepository, chatRepository, chatMemberRepository, messageDeletedPublisher)
	editMessageHandler := edit_message.NewHandler(messageRepository, messageEditedPublisher)
	updateChatHandler := update_chat.NewHandler(chatRepository, chatMemberRepository)
	listChatMessagesHandler := list_chat_messages.NewHandler(messageRepository, chatRepository, chatMemberRepository)
	listChatsHandler := list_chats.NewHandler(chatRepository, chatMemberRepository, messageRepository)
	setLastSeenRevisionHandler := set_last_seen_revision.NewHandler(chatRepository, chatMemberRepository, lastSeenRevisionUpdatedPublisher)

	// --- Section: HTTP transport ---
	authorizeChatAccessHttpHandler := transport_http.NewAuthorizeChatAccessHandler(authorizeChatAccessHandler)
	createDirectChatHttpHandler := transport_http.NewCreateDirectChatHandler(createDirectChatHandler)
	createGroupChatHttpHandler := transport_http.NewCreateGroupChatHandler(createGroupChatHandler)
	addChatMembersHttpHandler := transport_http.NewAddChatMembersHandler(addChatMembersHandler)
	kickChatMemberHttpHandler := transport_http.NewKickChatMemberHandler(kickChatMemberHandler)
	createMessageHttpHandler := transport_http.NewCreateMessageHandler(createMessageHandler)
	deleteMessageHttpHandler := transport_http.NewDeleteMessageHandler(deleteMessageHandler)
	editMessageHttpHandler := transport_http.NewEditMessageHandler(editMessageHandler)
	updateChatHttpHandler := transport_http.NewUpdateChatHandler(updateChatHandler)
	listChatMessagesHttpHandler := transport_http.NewListChatMessagesHandler(listChatMessagesHandler)
	listChatsHttpHandler := transport_http.NewListChatsHandler(listChatsHandler)
	setLastSeenRevisionHttpHandler := transport_http.NewSetLastSeenRevisionHandler(setLastSeenRevisionHandler)

	// --- Section: HTTP routes ---
	webServer := http.NewServeMux()
	webServer.HandleFunc("POST   /api/v1/message/chats/direct", middleware.RequireAuthenticatedUser(middleware.RejectLargeRequest(2048, middleware.WriteJsonResponse(createDirectChatHttpHandler.Handle))))
	webServer.HandleFunc("POST   /api/v1/message/chats/group", middleware.RequireAuthenticatedUser(middleware.RejectLargeRequest(2048, middleware.WriteJsonResponse(createGroupChatHttpHandler.Handle))))
	webServer.HandleFunc("POST   /api/v1/message/chats/{chatUuid}/members", middleware.RequireAuthenticatedUser(middleware.RejectLargeRequest(2048, middleware.WriteJsonResponse(addChatMembersHttpHandler.Handle))))
	webServer.HandleFunc("DELETE /api/v1/message/chats/{chatUuid}/members/{userUuid}", middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(kickChatMemberHttpHandler.Handle)))
	webServer.HandleFunc("POST   /api/v1/message/chats/{chatUuid}/authorize-access", middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(authorizeChatAccessHttpHandler.Handle)))
	webServer.HandleFunc("GET    /api/v1/message/chats", middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(listChatsHttpHandler.Handle)))
	webServer.HandleFunc("POST   /api/v1/message/chats/{chatUuid}/last-seen-revision", middleware.RequireAuthenticatedUser(middleware.RejectLargeRequest(256, middleware.WriteJsonResponse(setLastSeenRevisionHttpHandler.Handle))))
	webServer.HandleFunc("POST   /api/v1/message/chats/{chatUuid}/messages", middleware.RequireAuthenticatedUser(middleware.RejectLargeRequest(4096, middleware.WriteJsonResponse(createMessageHttpHandler.Handle))))
	webServer.HandleFunc("GET    /api/v1/message/chats/{chatUuid}/messages", middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(listChatMessagesHttpHandler.Handle)))
	webServer.HandleFunc("PUT    /api/v1/message/chats/{chatUuid}", middleware.RequireAuthenticatedUser(middleware.RejectLargeRequest(2048, middleware.WriteJsonResponse(updateChatHttpHandler.Handle))))
	webServer.HandleFunc("PUT    /api/v1/message/messages/{messageUuid}", middleware.RequireAuthenticatedUser(middleware.RejectLargeRequest(4096, middleware.WriteJsonResponse(editMessageHttpHandler.Handle))))
	webServer.HandleFunc("DELETE /api/v1/message/messages/{messageUuid}", middleware.RequireAuthenticatedUser(middleware.WriteJsonResponse(deleteMessageHttpHandler.Handle)))

	// --- Section: HTTP server ---
	address := fmt.Sprintf("0.0.0.0:%d", configuration.App.Port)

	log.Printf("message service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}
