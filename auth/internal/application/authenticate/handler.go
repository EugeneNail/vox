package authenticate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/EugeneNail/vox/auth/internal/application"
	"github.com/EugeneNail/vox/auth/internal/domain"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/validation"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/validation/rules"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials is returned when the provided credentials do not match a user.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Handler authenticates users and issues login tokens.
type Handler struct {
	repository domain.UserRepository
}

// Query contains the input required to authenticate a user.
type Query struct {
	Email    string
	Password string
}

// NewHandler constructs an authenticate handler with its dependencies.
func NewHandler(repository domain.UserRepository) *Handler {
	return &Handler{
		repository: repository,
	}
}

// Handle validates input, verifies credentials, and issues access tokens.
func (handler *Handler) Handle(ctx context.Context, query Query) (string, string, error) {
	email := strings.ToLower(strings.TrimSpace(query.Email))

	validator := validation.NewValidator(map[string]any{
		"email":    email,
		"password": query.Password,
	}, map[string][]rules.Rule{
		"email":    {rules.Required(), rules.Regex(rules.EmailPattern)},
		"password": {rules.Required()},
	})

	if err := validator.Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return "", "", validationError
		}

		return "", "", fmt.Errorf("validating authenticate query: %w", err)
	}

	user, err := handler.repository.FindByEmail(ctx, email)
	if err != nil {
		return "", "", fmt.Errorf("finding user by email %q: %w", email, err)
	}

	if user == nil {
		return "", "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(query.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", "", ErrInvalidCredentials
		}

		return "", "", fmt.Errorf("comparing password for user %q: %w", email, err)
	}

	now := time.Now().UTC()
	loginToken, err := application.GenerateToken(application.LoginTokenType, user.Uuid.String(), now.Add(application.LoginTokenTTL))
	if err != nil {
		return "", "", fmt.Errorf("generating login token for user %q: %w", email, err)
	}

	refreshToken, err := application.GenerateToken(application.RefreshTokenType, user.Uuid.String(), now.Add(application.RefreshTokenTTL))
	if err != nil {
		return "", "", fmt.Errorf("generating refresh token for user %q: %w", email, err)
	}

	return loginToken, refreshToken, nil
}
