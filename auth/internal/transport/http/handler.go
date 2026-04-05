package http

import (
	"github.com/EugeneNail/vox/auth/internal/application/authenticate"
	"github.com/EugeneNail/vox/auth/internal/application/create_user"
)

type Handler struct {
	createUserHandler   *create_user.Handler
	authenticateHandler *authenticate.Handler
}

// NewHandler constructs a shared HTTP handler for auth routes.
func NewHandler(createUserHandler *create_user.Handler, authenticateHandler *authenticate.Handler) *Handler {
	return &Handler{
		createUserHandler:   createUserHandler,
		authenticateHandler: authenticateHandler,
	}
}
