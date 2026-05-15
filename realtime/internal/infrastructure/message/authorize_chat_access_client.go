package message

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

var ErrChatNotFound = errors.New("chat not found")
var ErrChatAccessDenied = errors.New("chat access denied")

// AuthorizeChatAccessClient authorizes chat access through the message service.
type AuthorizeChatAccessClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAuthorizeChatAccessClient constructs a message-service authorization client.
func NewAuthorizeChatAccessClient(baseURL string) *AuthorizeChatAccessClient {
	return &AuthorizeChatAccessClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{},
	}
}

// Authorize checks whether the bearer token can access the chat.
func (client *AuthorizeChatAccessClient) Authorize(ctx context.Context, chatUuid uuid.UUID, loginToken string) error {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/message/chats/%s/authorize-access", client.baseURL, chatUuid),
		nil,
	)
	if err != nil {
		return fmt.Errorf("building authorize chat access request for chat %q: %w", chatUuid, err)
	}

	request.Header.Set("Authorization", "Bearer "+loginToken)

	response, err := client.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("performing authorize chat access request for chat %q: %w", chatUuid, err)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return ErrChatNotFound
	case http.StatusForbidden:
		return ErrChatAccessDenied
	default:
		return fmt.Errorf("authorizing chat access for chat %q returned status %d", chatUuid, response.StatusCode)
	}
}
