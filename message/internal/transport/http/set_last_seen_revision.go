package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/EugeneNail/vox/lib-common/authentication"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/application/usecases/set_last_seen_revision"
	"github.com/google/uuid"
)

type setLastSeenRevisionPayload struct {
	Revision int64 `json:"revision"`
}

type SetLastSeenRevisionHandler struct {
	usecase *set_last_seen_revision.Handler
}

func NewSetLastSeenRevisionHandler(usecase *set_last_seen_revision.Handler) *SetLastSeenRevisionHandler {
	return &SetLastSeenRevisionHandler{
		usecase: usecase,
	}
}

// Handle decodes the request and calls the use-case.
func (handler *SetLastSeenRevisionHandler) Handle(request *http.Request) (int, any) {
	chatUuid, err := uuid.Parse(strings.TrimSpace(request.PathValue("chatUuid")))
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("parsing chat uuid %q: %w", request.PathValue("chatUuid"), err)
	}

	var payload setLastSeenRevisionPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	userUuid, ok := authentication.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	if err := handler.usecase.Handle(request.Context(), set_last_seen_revision.Command{
		ChatUuid: chatUuid,
		UserUuid: userUuid,
		Revision: payload.Revision,
	}); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		if errors.Is(err, set_last_seen_revision.ErrChatNotFound) {
			return http.StatusNotFound, fmt.Errorf("chat %q not found: %w", chatUuid, err)
		}

		if errors.Is(err, set_last_seen_revision.ErrChatAccessDenied) {
			return http.StatusForbidden, fmt.Errorf("access to chat %q denied for user %q: %w", chatUuid, userUuid, err)
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the SetLastSeenRevision usecase: %w", err)
	}

	return http.StatusOK, nil
}
