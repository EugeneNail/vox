package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_message"
	message_middleware "github.com/EugeneNail/vox/message/internal/infrastructure/http/middleware"
	"github.com/google/uuid"
)

type createMessagePayload struct {
	Text        string   `json:"text"`
	Attachments []string `json:"attachments"`
}

type CreateMessageHandler struct {
	usecase *create_message.Handler
}

func NewCreateMessageHandler(usecase *create_message.Handler) *CreateMessageHandler {
	return &CreateMessageHandler{
		usecase: usecase,
	}
}

// CreateMessage decodes the request and calls the use-case.
func (handler *CreateMessageHandler) Handle(request *http.Request) (int, any) {
	chatUuid, err := uuid.Parse(request.PathValue("chatUuid"))
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("parsing chat uuid %q: %w", request.PathValue("chatUuid"), err)
	}

	var payload createMessagePayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	// TODO
	// Extract into 'authentication' package
	userUuid, ok := message_middleware.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	messageUuid, err := handler.usecase.Handle(request.Context(), create_message.Command{
		ChatUuid:    chatUuid,
		UserUuid:    userUuid,
		Text:        payload.Text,
		Attachments: payload.Attachments,
	})
	if err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		if errors.Is(err, create_message.ErrChatNotFound) {
			return http.StatusNotFound, fmt.Errorf("chat %q not found: %w", chatUuid, err)
		}

		if errors.Is(err, create_message.ErrChatAccessDenied) {
			return http.StatusForbidden, fmt.Errorf("access to chat %q denied for user %q: %w", chatUuid, userUuid, err)
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the CreateMessage usecase: %w", err)
	}

	return http.StatusCreated, messageUuid
}
