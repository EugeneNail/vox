package authenticate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/EugeneNail/vox/auth/internal/domain"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/validation"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/validation/rules"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// tempSalt signs JWTs until a dedicated secret is introduced.
const tempSalt = "TEMP_SALT"

// loginTokenTTL defines how long the login token stays valid.
const loginTokenTTL = 15 * time.Minute

// refreshTokenTTL defines how long the refresh token stays valid.
const refreshTokenTTL = 7 * 24 * time.Hour

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
	loginToken, err := generateToken("login-token", user.Email, now.Add(loginTokenTTL))
	if err != nil {
		return "", "", fmt.Errorf("generating login token for user %q: %w", email, err)
	}

	refreshToken, err := generateToken("refresh-token", user.Email, now.Add(refreshTokenTTL))
	if err != nil {
		return "", "", fmt.Errorf("generating refresh token for user %q: %w", email, err)
	}

	return loginToken, refreshToken, nil
}

// generateToken signs a JWT for the given token type and expiration time.
func generateToken(tokenType string, email string, expiresAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  email,
		"type": tokenType,
		"iat":  time.Now().UTC().Unix(),
		"exp":  expiresAt.Unix(),
	})

	signedToken, err := token.SignedString([]byte(tempSalt))
	if err != nil {
		return "", fmt.Errorf("signing jwt token: %w", err)
	}

	return signedToken, nil
}
