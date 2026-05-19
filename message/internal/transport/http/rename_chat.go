package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/EugeneNail/vox/lib-common/authentication"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/application/usecases/rename_chat"
	"github.com/google/uuid"
)

type renameChatPayload struct {
	Name *string `json:"name"`
}

type RenameChatHandler struct {
	usecase *rename_chat.Handler
}

func NewRenameChatHandler(usecase *rename_chat.Handler) *RenameChatHandler {
	return &RenameChatHandler{
		usecase: usecase,
	}
}

// Handle decodes the request and calls the use-case.
func (handler *RenameChatHandler) Handle(request *http.Request) (int, any) {
	chatUuid, err := uuid.Parse(strings.TrimSpace(request.PathValue("chatUuid")))
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("parsing chat uuid %q: %w", request.PathValue("chatUuid"), err)
	}

	var payload renameChatPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	userUuid, ok := authentication.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	if err := handler.usecase.Handle(request.Context(), rename_chat.Command{
		ChatUuid: chatUuid,
		UserUuid: userUuid,
		Name:     payload.Name,
	}); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		if errors.Is(err, rename_chat.ErrChatNotFound) {
			return http.StatusNotFound, fmt.Errorf("chat %q not found: %w", chatUuid, err)
		}

		if errors.Is(err, rename_chat.ErrChatAccessDenied) {
			return http.StatusForbidden, fmt.Errorf("access to chat %q denied for user %q: %w", chatUuid, userUuid, err)
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the RenameChat usecase: %w", err)
	}

	return http.StatusOK, nil
}
