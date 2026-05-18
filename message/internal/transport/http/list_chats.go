package http

import (
	"fmt"
	"net/http"

	"github.com/EugeneNail/vox/lib-common/authentication"
	"github.com/EugeneNail/vox/message/internal/application/usecases/list_chats"
	"github.com/EugeneNail/vox/message/internal/domain"
	"github.com/EugeneNail/vox/message/internal/transport/http/resource"
)

type ListChatsHandler struct {
	usecase *list_chats.Handler
}

func NewListChatsHandler(usecase *list_chats.Handler) *ListChatsHandler {
	return &ListChatsHandler{
		usecase: usecase,
	}
}

// Handle calls the use-case and returns chats available to the user.
func (handler *ListChatsHandler) Handle(request *http.Request) (int, any) {
	userUuid, ok := authentication.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	results, err := handler.usecase.Handle(request.Context(), list_chats.Query{UserUuid: userUuid})
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("handling the ListChats usecase: %w", err)
	}

	resources := make([]resource.Chat, 0, len(results))
	for _, result := range results {
		chat := result.Chat
		resources = append(resources, resource.Chat{
			Uuid:                        chat.Uuid,
			Name:                        chat.Name,
			Avatar:                      chat.Avatar,
			ChatType:                    chat.ChatType,
			Revision:                    chat.Revision,
			CreatedByUserUuid:           chat.CreatedByUserUuid,
			MemberUuids:                 result.MemberUuids,
			CurrentUserRole:             result.CurrentUserRole,
			CurrentUserLastSeenRevision: result.CurrentUserLastSeenRevision,
			LastMessage:                 toMessageResource(result.LastMessage),
			CreatedAt:                   chat.CreatedAt,
			UpdatedAt:                   chat.UpdatedAt,
		})
	}

	return http.StatusOK, resources
}

func toMessageResource(message *domain.Message) *resource.Message {
	if message == nil {
		return nil
	}

	attachments := make([]resource.Attachment, 0, len(message.Attachments))
	for _, attachment := range message.Attachments {
		attachments = append(attachments, resource.Attachment{
			Uuid: attachment.Uuid,
			Name: attachment.Name,
		})
	}

	return &resource.Message{
		Uuid:        message.Uuid,
		ChatUuid:    message.ChatUuid,
		UserUuid:    message.UserUuid,
		Revision:    message.Revision,
		Text:        message.Text,
		Attachments: attachments,
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   message.UpdatedAt,
	}
}
