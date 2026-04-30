package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/EugeneNail/vox/lib-common/authentication"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_direct_chat"
	"github.com/google/uuid"
)

type createChatPayload struct {
	InterlocutorUuid string `json:"interlocutorUuid"`
}

type CreateDirectChatHandler struct {
	usecase *create_direct_chat.Handler
}

func NewCreateDirectChatHandler(usecase *create_direct_chat.Handler) *CreateDirectChatHandler {
	return &CreateDirectChatHandler{
		usecase: usecase,
	}
}

// Handle decodes the request and calls the use-case.
func (handler *CreateDirectChatHandler) Handle(request *http.Request) (int, any) {
	var payload createChatPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	interlocutorUuid, err := uuid.Parse(strings.TrimSpace(payload.InterlocutorUuid))
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("parsing interlocutor uuid %q: %w", payload.InterlocutorUuid, err)
	}

	userUuid, ok := authentication.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	chatUuid, err := handler.usecase.Handle(request.Context(), create_direct_chat.Command{
		CreatorUuid:      userUuid,
		InterlocutorUuid: interlocutorUuid,
	})
	if err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the CreateDirectChat usecase: %w", err)
	}

	return http.StatusCreated, chatUuid
}
