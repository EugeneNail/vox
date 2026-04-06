package http

import (
	"fmt"
	"net/http"

	"github.com/EugeneNail/vox/message/internal/application/usecases/list_direct_chats"
	message_middleware "github.com/EugeneNail/vox/message/internal/infrastructure/http/middleware"
	"github.com/EugeneNail/vox/message/internal/transport/http/resource"
)

// ListDirectChats calls the use-case and returns direct chats available to the user.
func (handler *Handler) ListDirectChats(request *http.Request) (int, any) {
	userUuid, ok := message_middleware.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	chats, err := handler.listDirectChatsHandler.Handle(request.Context(), list_direct_chats.Query{UserUuid: userUuid})
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("handling the ListDirectChats usecase: %w", err)
	}

	resources := make([]resource.Chat, 0, len(chats))
	for _, chat := range chats {
		resources = append(resources, resource.Chat{
			Uuid: chat.Uuid,
			Name: chat.Name,
		})
	}

	return http.StatusOK, resources
}
