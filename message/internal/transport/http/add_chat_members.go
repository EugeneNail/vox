package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/EugeneNail/vox/lib-common/authentication"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/application/usecases/add_chat_members"
	"github.com/google/uuid"
)

type addChatMembersPayload struct {
	MemberUuids []string `json:"memberUuids"`
}

type AddChatMembersHandler struct {
	usecase *add_chat_members.Handler
}

func NewAddChatMembersHandler(usecase *add_chat_members.Handler) *AddChatMembersHandler {
	return &AddChatMembersHandler{
		usecase: usecase,
	}
}

// Handle decodes the request and calls the use-case.
func (handler *AddChatMembersHandler) Handle(request *http.Request) (int, any) {
	chatUuid, err := uuid.Parse(strings.TrimSpace(request.PathValue("chatUuid")))
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("parsing chat uuid %q: %w", request.PathValue("chatUuid"), err)
	}

	var payload addChatMembersPayload
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

	if err := handler.usecase.Handle(request.Context(), add_chat_members.Command{
		ChatUuid:    chatUuid,
		UserUuid:    userUuid,
		MemberUuids: memberUuids,
	}); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		if errors.Is(err, add_chat_members.ErrChatNotFound) {
			return http.StatusNotFound, fmt.Errorf("chat %q not found: %w", chatUuid, err)
		}

		if errors.Is(err, add_chat_members.ErrChatAccessDenied) {
			return http.StatusForbidden, fmt.Errorf("access to chat %q denied for user %q: %w", chatUuid, userUuid, err)
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the AddChatMembers usecase: %w", err)
	}

	return http.StatusOK, nil
}
