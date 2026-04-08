package http

import (
	"fmt"
	"net/http"

	"github.com/EugeneNail/vox/message/internal/application/usecases/list_chats"
	message_middleware "github.com/EugeneNail/vox/message/internal/infrastructure/http/middleware"
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
	userUuid, ok := message_middleware.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	chats, err := handler.usecase.Handle(request.Context(), list_chats.Query{UserUuid: userUuid})
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("handling the ListChats usecase: %w", err)
	}

	resources := make([]resource.Chat, 0, len(chats))
	for _, chat := range chats {
		resources = append(resources, resource.Chat{
			Uuid:              chat.Uuid,
			Name:              chat.Name,
			Avatar:            chat.Avatar,
			CreatedByUserUuid: chat.CreatedByUserUuid,
			MemberUuids:       chat.MemberUuids,
			CreatedAt:         chat.CreatedAt,
			UpdatedAt:         chat.UpdatedAt,
		})
	}

	return http.StatusOK, resources
}
