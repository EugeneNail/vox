package refresh

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/EugeneNail/vox/auth/internal/application"
	"github.com/EugeneNail/vox/auth/internal/domain"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/validation"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/validation/rules"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ErrInvalidRefreshToken is returned when the provided refresh token is invalid.
var ErrInvalidRefreshToken = errors.New("invalid refresh token")

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
		"refreshToken": {rules.Required()},
	})

	if err := validator.Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return "", validationError
		}

		return "", fmt.Errorf("validating refresh query: %w", err)
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(query.RefreshToken, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidRefreshToken
		}

		return []byte(application.TokenSalt), nil
	})
	if err != nil {
		return "", ErrInvalidRefreshToken
	}

	if !token.Valid {
		return "", ErrInvalidRefreshToken
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != application.RefreshTokenType {
		return "", ErrInvalidRefreshToken
	}

	userUuidString, ok := claims["sub"].(string)
	if !ok || userUuidString == "" {
		return "", ErrInvalidRefreshToken
	}

	userUuid, err := uuid.Parse(userUuidString)
	if err != nil {
		return "", ErrInvalidRefreshToken
	}

	user, err := handler.repository.FindByUuid(ctx, userUuid)
	if err != nil {
		return "", fmt.Errorf("finding user by uuid %q: %w", userUuid, err)
	}

	if user == nil {
		return "", ErrInvalidRefreshToken
	}

	now := time.Now().UTC()
	loginToken, err := application.GenerateToken(application.LoginTokenType, user.Uuid.String(), now.Add(application.LoginTokenTTL))
	if err != nil {
		return "", fmt.Errorf("generating login token for user %q: %w", user.Uuid, err)
	}

	return loginToken, nil
}
