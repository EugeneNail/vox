package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/message/internal/application/usecases/list_chat_messages"
	message_middleware "github.com/EugeneNail/vox/message/internal/infrastructure/http/middleware"
	"github.com/EugeneNail/vox/message/internal/transport/http/resource"
	"github.com/google/uuid"
)

const defaultChatMessagesLength = 250
const maxChatMessagesLength = 250

type ListChatMessagesHandler struct {
	usecase *list_chat_messages.Handler
}

func NewListChatMessagesHandler(usecase *list_chat_messages.Handler) *ListChatMessagesHandler {
	return &ListChatMessagesHandler{
		usecase: usecase,
	}
}

// ListChatMessages validates transport input and calls the use-case.
func (handler *ListChatMessagesHandler) Handle(request *http.Request) (int, any) {
	chatUuid, err := uuid.Parse(strings.TrimSpace(request.PathValue("chatUuid")))
	if err != nil {
		validationError := validation.NewError()
		validationError.AddViolation("chatUuid", "Must be a valid UUID")
		return http.StatusBadRequest, validationError.Violations()
	}

	length, err := parseChatMessagesLength(request)
	if err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusBadRequest, validationError.Violations()
		}

		return http.StatusInternalServerError, fmt.Errorf("parsing chat messages length: %w", err)
	}

	userUuid, ok := message_middleware.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	messages, err := handler.usecase.Handle(request.Context(), list_chat_messages.Query{
		ChatUuid: chatUuid,
		UserUuid: userUuid,
		Length:   length,
	})
	if err != nil {
		if errors.Is(err, list_chat_messages.ErrChatNotFound) {
			return http.StatusNotFound, fmt.Errorf("chat %q not found: %w", chatUuid, err)
		}

		if errors.Is(err, list_chat_messages.ErrChatAccessDenied) {
			return http.StatusForbidden, fmt.Errorf("access to chat %q denied for user %q: %w", chatUuid, userUuid, err)
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the ListChatMessages usecase: %w", err)
	}

	resources := make([]resource.Message, 0, len(messages))
	for _, message := range messages {
		attachments := make([]resource.Attachment, 0, len(message.Attachments))
		for _, attachment := range message.Attachments {
			attachments = append(attachments, resource.Attachment{
				Uuid: attachment.Uuid,
				Name: attachment.Name,
			})
		}

		resources = append(resources, resource.Message{
			Uuid:        message.Uuid,
			ChatUuid:    message.ChatUuid,
			UserUuid:    message.UserUuid,
			Text:        message.Text,
			Attachments: attachments,
			CreatedAt:   message.CreatedAt,
			UpdatedAt:   message.UpdatedAt,
		})
	}

	return http.StatusOK, resources
}

func parseChatMessagesLength(request *http.Request) (int, error) {
	rawLength := strings.TrimSpace(request.URL.Query().Get("length"))
	if rawLength == "" {
		return defaultChatMessagesLength, nil
	}

	length, err := strconv.Atoi(rawLength)
	if err != nil {
		validationError := validation.NewError()
		validationError.AddViolation("length", "Must be a valid integer")
		return 0, validationError
	}

	if length < 1 || length > maxChatMessagesLength {
		validationError := validation.NewError()
		validationError.AddViolation("length", fmt.Sprintf("Must be between 1 and %d", maxChatMessagesLength))
		return 0, validationError
	}

	return length, nil
}
