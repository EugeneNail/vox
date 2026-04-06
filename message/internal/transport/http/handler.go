package http

import "github.com/EugeneNail/vox/message/internal/application/usecases/create_message"

// Handler groups HTTP route handlers for the message service.
type Handler struct {
	createMessageHandler *create_message.Handler
}

// NewHandler constructs a shared HTTP handler for message routes.
func NewHandler(createMessageHandler *create_message.Handler) *Handler {
	return &Handler{
		createMessageHandler: createMessageHandler,
	}
}
