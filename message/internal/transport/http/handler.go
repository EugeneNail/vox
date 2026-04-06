package http

import (
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_direct_chat"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_message"
	"github.com/EugeneNail/vox/message/internal/application/usecases/list_direct_chats"
)

// Handler groups HTTP route handlers for the message service.
type Handler struct {
	createDirectChatHandler *create_direct_chat.Handler
	createMessageHandler    *create_message.Handler
	listDirectChatsHandler  *list_direct_chats.Handler
}

// NewHandler constructs a shared HTTP handler for message routes.
func NewHandler(createDirectChatHandler *create_direct_chat.Handler, createMessageHandler *create_message.Handler, listDirectChatsHandler *list_direct_chats.Handler) *Handler {
	return &Handler{
		createDirectChatHandler: createDirectChatHandler,
		createMessageHandler:    createMessageHandler,
		listDirectChatsHandler:  listDirectChatsHandler,
	}
}
