package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/lib-common/validation/rules"
	"github.com/EugeneNail/vox/message/internal/application/usecases/edit_message"
	message_middleware "github.com/EugeneNail/vox/message/internal/infrastructure/http/middleware"
	"github.com/google/uuid"
)

type editMessagePayload struct {
	Text string `json:"text"`
}

type EditMessageHandler struct {
	usecase *edit_message.Handler
}

func NewEditMessageHandler(usecase *edit_message.Handler) *EditMessageHandler {
	return &EditMessageHandler{
		usecase: usecase,
	}
}

// EditMessage decodes the request, applies transport validation, and calls the use-case.
func (handler *EditMessageHandler) Handle(request *http.Request) (int, any) {
	messageUuid, err := uuid.Parse(strings.TrimSpace(request.PathValue("messageUuid")))
	if err != nil {
		validationError := validation.NewError()
		validationError.AddViolation("messageUuid", "Must be a valid UUID")
		return http.StatusBadRequest, validationError.Violations()
	}

	var payload editMessagePayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	validator := validation.NewValidator(map[string]any{
		"text": payload.Text,
	}, map[string][]rules.Rule{
		"text": {rules.Required(), rules.Max(3000)},
	})

	if err := validator.Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusBadRequest, validationError.Violations()
		}

		return http.StatusInternalServerError, fmt.Errorf("validating edit message payload: %w", err)
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
	}); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		if errors.Is(err, edit_message.ErrMessageNotFound) || errors.Is(err, edit_message.ErrChatNotFound) {
			return http.StatusNotFound, err
		}

		if errors.Is(err, edit_message.ErrMessageAccessDenied) || errors.Is(err, edit_message.ErrChatAccessDenied) {
			return http.StatusForbidden, err
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the EditMessage usecase: %w", err)
	}

	return http.StatusOK, messageUuid
}
