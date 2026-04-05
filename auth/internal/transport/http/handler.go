package http

import (
	"github.com/EugeneNail/vox/auth/internal/application/usecases/authenticate"
	"github.com/EugeneNail/vox/auth/internal/application/usecases/create_user"
	"github.com/EugeneNail/vox/auth/internal/application/usecases/refresh"
)

type Handler struct {
	createUserHandler   *create_user.Handler
	authenticateHandler *authenticate.Handler
	refreshHandler      *refresh.Handler
}

// NewHandler constructs a shared HTTP handler for auth routes.
func NewHandler(createUserHandler *create_user.Handler, authenticateHandler *authenticate.Handler, refreshHandler *refresh.Handler) *Handler {
	return &Handler{
		createUserHandler:   createUserHandler,
		authenticateHandler: authenticateHandler,
		refreshHandler:      refreshHandler,
	}
}
