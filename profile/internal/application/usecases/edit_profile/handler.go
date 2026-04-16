package edit_profile

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

var ErrProfileNotFound = errors.New("profile not found")

// Handler edits public profiles through the edit_profile use-case.
type Handler struct {
	repository domain.ProfileRepository
}

// Command contains the input required to edit a profile.
type Command struct {
	UserUuid uuid.UUID
	Name     string
	Avatar   *string
}

// NewHandler constructs an edit_profile handler with its dependencies.
func NewHandler(repository domain.ProfileRepository) *Handler {
	return &Handler{
		repository: repository,
	}
}

// Handle validates input and updates an existing profile.
func (handler *Handler) Handle(ctx context.Context, command Command) (*domain.Profile, error) {
	command.Name = strings.TrimSpace(command.Name)

	fields := map[string]any{
		"userUuid": command.UserUuid,
		"name":     command.Name,
	}

	rls := map[string][]rules.Rule{
		"userUuid": {rules.Required()},
		"name":     {rules.Required(), rules.Min(2), rules.Max(64), rules.Regex(rules.SlugWithSpacesPattern)},
	}

	if command.Avatar != nil {
		fields["avatar"] = *command.Avatar
		rls["avatar"] = []rules.Rule{rules.Required(), rules.Max(100), rules.Regex(domain.AvatarPattern)}
	}

	if err := validation.NewValidator(fields, rls).Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return nil, validationError
		}

		return nil, fmt.Errorf("validating edit profile command for user %q: %w", command.UserUuid, err)
	}

	profile, err := handler.repository.FindByUserUuid(ctx, command.UserUuid)
	if err != nil {
		return nil, fmt.Errorf("finding profile for user %q: %w", command.UserUuid, err)
	}

	if profile == nil {
		return nil, ErrProfileNotFound
	}

	profile.Name = command.Name
	profile.UpdatedAt = time.Now().UTC()
	if command.Avatar != nil {
		profile.Avatar = command.Avatar
	}

	if err := handler.repository.Update(ctx, *profile); err != nil {
		return nil, fmt.Errorf("updating profile for user %q: %w", command.UserUuid, err)
	}

	return profile, nil
}
