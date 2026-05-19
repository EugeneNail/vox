package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/EugeneNail/vox/lib-common/authentication"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/application/usecases/change_chat_avatar"
	"github.com/google/uuid"
)

type changeChatAvatarPayload struct {
	Avatar *string `json:"avatar"`
}

type ChangeChatAvatarHandler struct {
	usecase *change_chat_avatar.Handler
}

func NewChangeChatAvatarHandler(usecase *change_chat_avatar.Handler) *ChangeChatAvatarHandler {
	return &ChangeChatAvatarHandler{
		usecase: usecase,
	}
}

// Handle decodes the request and calls the use-case.
func (handler *ChangeChatAvatarHandler) Handle(request *http.Request) (int, any) {
	chatUuid, err := uuid.Parse(strings.TrimSpace(request.PathValue("chatUuid")))
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("parsing chat uuid %q: %w", request.PathValue("chatUuid"), err)
	}

	var payload changeChatAvatarPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	userUuid, ok := authentication.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	if err := handler.usecase.Handle(request.Context(), change_chat_avatar.Command{
		ChatUuid: chatUuid,
		UserUuid: userUuid,
		Avatar:   payload.Avatar,
	}); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		if errors.Is(err, change_chat_avatar.ErrChatNotFound) {
			return http.StatusNotFound, fmt.Errorf("chat %q not found: %w", chatUuid, err)
		}

		if errors.Is(err, change_chat_avatar.ErrChatAccessDenied) {
			return http.StatusForbidden, fmt.Errorf("access to chat %q denied for user %q: %w", chatUuid, userUuid, err)
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the ChangeChatAvatar usecase: %w", err)
	}

	return http.StatusOK, nil
}
