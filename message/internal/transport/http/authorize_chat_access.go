package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/EugeneNail/vox/lib-common/authentication"
	"github.com/EugeneNail/vox/lib-common/validation"
	authorize_chat_access "github.com/EugeneNail/vox/message/internal/application/usecases/authorize_chat_access"
	"github.com/google/uuid"
)

type AuthorizeChatAccessHandler struct {
	usecase *authorize_chat_access.Handler
}

func NewAuthorizeChatAccessHandler(usecase *authorize_chat_access.Handler) *AuthorizeChatAccessHandler {
	return &AuthorizeChatAccessHandler{
		usecase: usecase,
	}
}

// Handle authorizes chat access for the authenticated user.
func (handler *AuthorizeChatAccessHandler) Handle(request *http.Request) (int, any) {
	chatUuid, err := uuid.Parse(strings.TrimSpace(request.PathValue("chatUuid")))
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("parsing chat uuid %q: %w", request.PathValue("chatUuid"), err)
	}

	userUuid, ok := authentication.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	if err := handler.usecase.Handle(request.Context(), authorize_chat_access.Command{
		ChatUuid: chatUuid,
		UserUuid: userUuid,
	}); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		if errors.Is(err, authorize_chat_access.ErrChatNotFound) {
			return http.StatusNotFound, fmt.Errorf("chat %q not found: %w", chatUuid, err)
		}

		if errors.Is(err, authorize_chat_access.ErrChatAccessDenied) {
			return http.StatusForbidden, fmt.Errorf("access to chat %q denied for user %q: %w", chatUuid, userUuid, err)
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the AuthorizeChatAccess usecase: %w", err)
	}

	return http.StatusOK, nil
}
