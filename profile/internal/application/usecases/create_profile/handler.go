package create_profile

import (
	"context"
	"fmt"
	"strings"
	"time"

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
}

// NewHandler constructs a create_profile handler with its dependencies.
func NewHandler(repository domain.ProfileRepository) *Handler {
	return &Handler{
		repository: repository,
	}
}

// Handle creates a profile when it does not exist yet.
func (handler *Handler) Handle(ctx context.Context, command Command) error {
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
		Name:      "New User",
		Nickname:  buildDefaultNickname(command.UserUuid),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := handler.repository.Create(ctx, profile); err != nil {
		return fmt.Errorf("creating profile for user %q: %w", command.UserUuid, err)
	}

	return nil
}

func buildDefaultNickname(userUuid uuid.UUID) string {
	return fmt.Sprintf("user-%s", strings.ToLower(userUuid.String()[:8]))
}
