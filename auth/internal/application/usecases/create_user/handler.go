package create_user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/EugeneNail/vox/auth/internal/domain"
	"github.com/EugeneNail/vox/auth/internal/domain/events"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/lib-common/validation/rules"
	"github.com/google/uuid"
	"github.com/samborkent/uuidv7"
	"golang.org/x/crypto/bcrypt"
)

var ErrEmailAlreadyExists = errors.New("user with this email already exists")

// Handler creates users through the create_user use-case.
type Handler struct {
	repository           domain.UserRepository
	userCreatedPublisher events.UserCreatedPublisher
}

// Command contains the input required to create a user.
type Command struct {
	Name                 string
	Email                string
	Password             string
	PasswordConfirmation string
}

// NewHandler constructs a create_user handler with its dependencies.
func NewHandler(repository domain.UserRepository, userCreatedPublisher events.UserCreatedPublisher) *Handler {
	return &Handler{
		repository:           repository,
		userCreatedPublisher: userCreatedPublisher,
	}
}

// Handle validates input, checks uniqueness, and creates a new user.
func (handler *Handler) Handle(ctx context.Context, command Command) (uuid.UUID, error) {
	name := strings.TrimSpace(command.Name)
	email := strings.ToLower(strings.TrimSpace(command.Email))

	validator := validation.NewValidator(map[string]any{
		"name":                 name,
		"email":                email,
		"password":             command.Password,
		"passwordConfirmation": command.PasswordConfirmation,
	}, map[string][]rules.Rule{
		"name":                 {rules.Required(), rules.Min(2), rules.Max(64), rules.Regex(rules.SlugWithSpacesPattern)},
		"email":                {rules.Required(), rules.Regex(rules.EmailPattern)},
		"password":             {rules.Required(), rules.Min(8), rules.Password()},
		"passwordConfirmation": {rules.Required(), rules.Same("password")},
	})

	if err := validator.Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return uuid.Nil, validationError
		}

		return uuid.Nil, fmt.Errorf("validating create user command: %w", err)
	}

	existingUser, err := handler.repository.FindByEmail(ctx, email)
	if err != nil {
		return uuid.Nil, fmt.Errorf("finding user by email %q: %w", email, err)
	}

	if existingUser != nil {
		return uuid.Nil, ErrEmailAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(command.Password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hashing password: %w", err)
	}

	now := time.Now().UTC()
	user := domain.User{
		Uuid:      uuid.UUID(uuidv7.New()),
		Email:     email,
		Password:  string(passwordHash),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := handler.repository.Create(ctx, user); err != nil {
		return uuid.Nil, fmt.Errorf("creating user with email %q: %w", user.Email, err)
	}

	if err := handler.userCreatedPublisher.Publish(ctx, events.UserCreated{
		UserUuid:  user.Uuid,
		Name:      name,
		CreatedAt: user.CreatedAt,
	}); err != nil {
		return uuid.Nil, fmt.Errorf("publishing user created event for user %q: %w", user.Uuid, err)
	}

	return user.Uuid, nil
}
