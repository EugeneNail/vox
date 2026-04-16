package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/EugeneNail/vox/lib-common/authentication"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_chat"
	"github.com/google/uuid"
)

type createChatPayload struct {
	MemberUuids []string `json:"memberUuids"`
	Name        *string  `json:"name"`
	Avatar      *string  `json:"avatar"`
	IsPrivate   bool     `json:"isPrivate"`
}

type CreateChatHandler struct {
	usecase *create_chat.Handler
}

func NewCreateChatHandler(usecase *create_chat.Handler) *CreateChatHandler {
	return &CreateChatHandler{
		usecase: usecase,
	}
}

// Handle decodes the request and calls the use-case.
func (handler *CreateChatHandler) Handle(request *http.Request) (int, any) {
	var payload createChatPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	memberUuids := make([]uuid.UUID, 0, len(payload.MemberUuids))
	for index, rawMemberUuid := range payload.MemberUuids {
		memberUuid, err := uuid.Parse(strings.TrimSpace(rawMemberUuid))
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("parsing member uuid at index %d: %w", index, err)
		}

		memberUuids = append(memberUuids, memberUuid)
	}

	userUuid, ok := authentication.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	chatUuid, err := handler.usecase.Handle(request.Context(), create_chat.Command{
		CreatorUuid: userUuid,
		MemberUuids: memberUuids,
		Name:        payload.Name,
		Avatar:      payload.Avatar,
		IsPrivate:   payload.IsPrivate,
	})
	if err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the CreateChat usecase: %w", err)
	}

	return http.StatusCreated, chatUuid
}
