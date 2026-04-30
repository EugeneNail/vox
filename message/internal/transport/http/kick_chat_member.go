package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/EugeneNail/vox/lib-common/authentication"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/application/usecases/kick_chat_member"
	"github.com/google/uuid"
)

type KickChatMemberHandler struct {
	usecase *kick_chat_member.Handler
}

func NewKickChatMemberHandler(usecase *kick_chat_member.Handler) *KickChatMemberHandler {
	return &KickChatMemberHandler{
		usecase: usecase,
	}
}

// Handle decodes the request and calls the use-case.
func (handler *KickChatMemberHandler) Handle(request *http.Request) (int, any) {
	chatUuid, err := uuid.Parse(strings.TrimSpace(request.PathValue("chatUuid")))
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("parsing chat uuid %q: %w", request.PathValue("chatUuid"), err)
	}

	memberUuid, err := uuid.Parse(strings.TrimSpace(request.PathValue("userUuid")))
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("parsing user uuid %q: %w", request.PathValue("userUuid"), err)
	}

	userUuid, ok := authentication.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	if err := handler.usecase.Handle(request.Context(), kick_chat_member.Command{
		ChatUuid:   chatUuid,
		UserUuid:   userUuid,
		MemberUuid: memberUuid,
	}); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		if errors.Is(err, kick_chat_member.ErrChatNotFound) {
			return http.StatusNotFound, fmt.Errorf("chat %q not found: %w", chatUuid, err)
		}

		if errors.Is(err, kick_chat_member.ErrChatAccessDenied) {
			return http.StatusForbidden, fmt.Errorf("access to chat %q denied for user %q: %w", chatUuid, userUuid, err)
		}

		if errors.Is(err, kick_chat_member.ErrChatMemberNotFound) {
			return http.StatusNotFound, fmt.Errorf("member %q not found in chat %q: %w", memberUuid, chatUuid, err)
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the KickChatMember usecase: %w", err)
	}

	return http.StatusOK, nil
}
