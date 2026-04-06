package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_direct_chat"
	message_middleware "github.com/EugeneNail/vox/message/internal/infrastructure/http/middleware"
	"github.com/google/uuid"
)

type createDirectChatPayload struct {
	CompanionUuid string `json:"companionUuid"`
}

// CreateDirectChat decodes the request and calls the use-case.
func (handler *Handler) CreateDirectChat(request *http.Request) (int, any) {
	var payload createDirectChatPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	companionUuid, err := uuid.Parse(strings.TrimSpace(payload.CompanionUuid))
	if err != nil {
		validationError := validation.NewError()
		validationError.AddViolation("companionUuid", "Must be a valid UUID")
		return http.StatusBadRequest, validationError.Violations()
	}

	userUuid, ok := message_middleware.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	chatUuid, err := handler.createDirectChatHandler.Handle(request.Context(), create_direct_chat.Command{
		CompanionUuid: companionUuid,
		CreatorUuid:   userUuid,
	})
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("handling the CreateDirectChat usecase: %w", err)
	}

	return http.StatusCreated, chatUuid
}
