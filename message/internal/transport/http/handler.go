package http

import (
	"github.com/EugeneNail/vox/message/internal/application/usecases/authorize_direct_chat_updates"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_direct_chat"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_message"
	"github.com/EugeneNail/vox/message/internal/application/usecases/edit_message"
	"github.com/EugeneNail/vox/message/internal/application/usecases/list_chat_messages"
	"github.com/EugeneNail/vox/message/internal/application/usecases/list_direct_chats"
	websocket_infrastructure "github.com/EugeneNail/vox/message/internal/infrastructure/websocket"
)

// Handler groups HTTP route handlers for the message service.
type Handler struct {
	authorizeDirectChatUpdatesHandler *authorize_direct_chat_updates.Handler
	createDirectChatHandler           *create_direct_chat.Handler
	createMessageHandler              *create_message.Handler
	editMessageHandler                *edit_message.Handler
	listChatMessagesHandler           *list_chat_messages.Handler
	listDirectChatsHandler            *list_direct_chats.Handler
	connectionHub                     *websocket_infrastructure.ConnectionHub
	subscriptionRegistry              *websocket_infrastructure.ChatSubscriptionRegistry
}

// NewHandler constructs a shared HTTP handler for message routes.
func NewHandler(authorizeDirectChatUpdatesHandler *authorize_direct_chat_updates.Handler, createDirectChatHandler *create_direct_chat.Handler, createMessageHandler *create_message.Handler, editMessageHandler *edit_message.Handler, listChatMessagesHandler *list_chat_messages.Handler, listDirectChatsHandler *list_direct_chats.Handler, connectionHub *websocket_infrastructure.ConnectionHub, subscriptionRegistry *websocket_infrastructure.ChatSubscriptionRegistry) *Handler {
	return &Handler{
		authorizeDirectChatUpdatesHandler: authorizeDirectChatUpdatesHandler,
		createDirectChatHandler:           createDirectChatHandler,
		createMessageHandler:              createMessageHandler,
		editMessageHandler:                editMessageHandler,
		listChatMessagesHandler:           listChatMessagesHandler,
		listDirectChatsHandler:            listDirectChatsHandler,
		connectionHub:                     connectionHub,
		subscriptionRegistry:              subscriptionRegistry,
	}
}
