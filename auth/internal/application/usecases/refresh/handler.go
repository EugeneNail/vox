package refresh

import (
	"context"
	"errors"
	"fmt"

	"github.com/EugeneNail/vox/auth/internal/domain"
	"github.com/EugeneNail/vox/lib-common/authentication"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/lib-common/validation/rules"
)

// ErrInvalidToken is returned when the provided refresh token is invalid.
var ErrInvalidToken = errors.New("invalid token")

// Handler validates refresh tokens and issues a new login token.
type Handler struct {
	repository domain.UserRepository
}

// Query contains the input required to refresh tokens.
type Query struct {
	RefreshToken string
}

// NewHandler constructs a refresh handler with its dependencies.
func NewHandler(repository domain.UserRepository) *Handler {
	return &Handler{
		repository: repository,
	}
}

// Handle validates the refresh token and issues a new login token.
func (handler *Handler) Handle(ctx context.Context, query Query) (string, error) {
	validator := validation.NewValidator(map[string]any{
		"refreshToken": query.RefreshToken,
	}, map[string][]rules.Rule{
		"refreshToken": {rules.Required(), rules.Max(4096)},
	})

	if err := validator.Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return "", validationError
		}

		return "", fmt.Errorf("validating refresh query: %w", err)
	}

	userUuid, err := authentication.UserUuidFromRefreshToken(query.RefreshToken)
	if errors.Is(err, authentication.ErrInvalidToken) {
		return "", ErrInvalidToken
	}
	if err != nil {
		return "", fmt.Errorf("validating refresh token: %w", err)
	}

	user, err := handler.repository.FindByUuid(ctx, userUuid)
	if err != nil {
		return "", fmt.Errorf("finding user by uuid %q: %w", userUuid, err)
	}

	if user == nil {
		return "", ErrInvalidToken
	}

	loginToken, err := authentication.NewTokenSigner().NewLoginToken(user.Uuid.String())
	if err != nil {
		return "", fmt.Errorf("generating login token for user %q: %w", user.Uuid, err)
	}

	return loginToken, nil
}
