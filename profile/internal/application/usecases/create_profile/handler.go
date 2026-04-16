package create_profile

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/lib-common/validation/rules"
	"github.com/EugeneNail/vox/profile/internal/domain"
	"github.com/google/uuid"
)

// Handler creates public profiles from user-created events.
type Handler struct {
	repository domain.ProfileRepository
}

// Command contains the input required to create a profile.
type Command struct {
	UserUuid uuid.UUID
	Name     string
}

// NewHandler constructs a create_profile handler with its dependencies.
func NewHandler(repository domain.ProfileRepository) *Handler {
	return &Handler{
		repository: repository,
	}
}

// Handle creates a profile when it does not exist yet.
func (handler *Handler) Handle(ctx context.Context, command Command) error {
	command.Name = strings.TrimSpace(command.Name)

	validator := validation.NewValidator(map[string]any{
		"userUuid": command.UserUuid,
		"name":     command.Name,
	}, map[string][]rules.Rule{
		"userUuid": {rules.Required()},
		"name":     {rules.Required(), rules.Min(2), rules.Max(64), rules.Regex(rules.SlugWithSpacesPattern)},
	})

	if err := validator.Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return validationError
		}

		return fmt.Errorf("validating create profile command for user %q: %w", command.UserUuid, err)
	}

	existingProfile, err := handler.repository.FindByUserUuid(ctx, command.UserUuid)
	if err != nil {
		return fmt.Errorf("finding profile for user %q: %w", command.UserUuid, err)
	}

	if existingProfile != nil {
		return nil
	}

	now := time.Now().UTC()
	profile := domain.Profile{
		UserUuid:  command.UserUuid,
		Avatar:    nil,
		Name:      command.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := handler.repository.Create(ctx, profile); err != nil {
		return fmt.Errorf("creating profile for user %q: %w", command.UserUuid, err)
	}

	return nil
}
