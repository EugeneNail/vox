package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/application/usecases/edit_message"
	message_middleware "github.com/EugeneNail/vox/message/internal/infrastructure/http/middleware"
	"github.com/google/uuid"
)

type editMessagePayload struct {
	Text        string   `json:"text"`
	Attachments []string `json:"attachments"`
}

type EditMessageHandler struct {
	usecase *edit_message.Handler
}

func NewEditMessageHandler(usecase *edit_message.Handler) *EditMessageHandler {
	return &EditMessageHandler{
		usecase: usecase,
	}
}

// EditMessage decodes the request and calls the use-case.
func (handler *EditMessageHandler) Handle(request *http.Request) (int, any) {
	messageUuid, err := uuid.Parse(strings.TrimSpace(request.PathValue("messageUuid")))
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("parsing message uuid %q: %w", request.PathValue("messageUuid"), err)
	}

	var payload editMessagePayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	// TODO
	// Extract into 'authentication' package
	userUuid, ok := message_middleware.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	if err := handler.usecase.Handle(request.Context(), edit_message.Command{
		MessageUuid: messageUuid,
		UserUuid:    userUuid,
		Text:        payload.Text,
		Attachments: payload.Attachments,
	}); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		if errors.Is(err, edit_message.ErrMessageNotFound) || errors.Is(err, edit_message.ErrChatNotFound) {
			return http.StatusNotFound, fmt.Errorf("message %q or its chat not found: %w", messageUuid, err)
		}

		if errors.Is(err, edit_message.ErrMessageAccessDenied) || errors.Is(err, edit_message.ErrChatAccessDenied) {
			return http.StatusForbidden, fmt.Errorf("access denied for user %q editing message %q: %w", userUuid, messageUuid, err)
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the EditMessage usecase: %w", err)
	}

	return http.StatusOK, messageUuid
}
