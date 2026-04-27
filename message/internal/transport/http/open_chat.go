package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/EugeneNail/vox/lib-common/authentication"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/application/usecases/authorize_chat_updates"
	open_chat "github.com/EugeneNail/vox/message/internal/application/usecases/open_chat"
	"github.com/google/uuid"
)

type OpenChatHandler struct {
	usecase *open_chat.Handler
}

func NewOpenChatHandler(usecase *open_chat.Handler) *OpenChatHandler {
	return &OpenChatHandler{
		usecase: usecase,
	}
}

// Handle opens a chat for the authenticated user.
func (handler *OpenChatHandler) Handle(request *http.Request) (int, any) {
	chatUuid, err := uuid.Parse(strings.TrimSpace(request.PathValue("chatUuid")))
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("parsing chat uuid %q: %w", request.PathValue("chatUuid"), err)
	}

	userUuid, ok := authentication.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	if err := handler.usecase.Handle(request.Context(), open_chat.Command{
		ChatUuid: chatUuid,
		UserUuid: userUuid,
	}); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		if errors.Is(err, authorize_chat_updates.ErrChatNotFound) {
			return http.StatusNotFound, fmt.Errorf("chat %q not found: %w", chatUuid, err)
		}

		if errors.Is(err, authorize_chat_updates.ErrChatAccessDenied) {
			return http.StatusForbidden, fmt.Errorf("access to chat %q denied for user %q: %w", chatUuid, userUuid, err)
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the OpenChat usecase: %w", err)
	}

	return http.StatusOK, nil
}
