package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/EugeneNail/vox/lib-common/authentication"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/application/usecases/delete_message"
	"github.com/google/uuid"
)

type DeleteMessageHandler struct {
	usecase *delete_message.Handler
}

func NewDeleteMessageHandler(usecase *delete_message.Handler) *DeleteMessageHandler {
	return &DeleteMessageHandler{
		usecase: usecase,
	}
}

// DeleteMessage calls the use-case.
func (handler *DeleteMessageHandler) Handle(request *http.Request) (int, any) {
	messageUuid, err := uuid.Parse(strings.TrimSpace(request.PathValue("messageUuid")))
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("parsing message uuid %q: %w", request.PathValue("messageUuid"), err)
	}

	userUuid, ok := authentication.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	if err := handler.usecase.Handle(request.Context(), delete_message.Command{
		MessageUuid: messageUuid,
		UserUuid:    userUuid,
	}); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		if errors.Is(err, delete_message.ErrMessageNotFound) || errors.Is(err, delete_message.ErrChatNotFound) {
			return http.StatusNotFound, err
		}

		if errors.Is(err, delete_message.ErrMessageAccessDenied) || errors.Is(err, delete_message.ErrChatAccessDenied) {
			return http.StatusForbidden, err
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the DeleteMessage usecase: %w", err)
	}

	return http.StatusOK, messageUuid
}
