package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/EugeneNail/vox/lib-common/authentication"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_group_chat"
	"github.com/google/uuid"
)

type createGroupChatPayload struct {
	MemberUuids []string `json:"memberUuids"`
	Name        *string  `json:"name"`
	Avatar      *string  `json:"avatar"`
}

type CreateGroupChatHandler struct {
	usecase *create_group_chat.Handler
}

func NewCreateGroupChatHandler(usecase *create_group_chat.Handler) *CreateGroupChatHandler {
	return &CreateGroupChatHandler{
		usecase: usecase,
	}
}

// Handle decodes the request and calls the use-case.
func (handler *CreateGroupChatHandler) Handle(request *http.Request) (int, any) {
	var payload createGroupChatPayload
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

	chatUuid, err := handler.usecase.Handle(request.Context(), create_group_chat.Command{
		CreatorUuid: userUuid,
		MemberUuids: memberUuids,
		Name:        payload.Name,
		Avatar:      payload.Avatar,
	})
	if err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the CreateGroupChat usecase: %w", err)
	}

	return http.StatusCreated, chatUuid
}
