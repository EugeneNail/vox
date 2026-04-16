package search_profiles

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/lib-common/validation/rules"
	"github.com/EugeneNail/vox/profile/internal/domain"
	"github.com/google/uuid"
)

type Handler struct {
	repository domain.ProfileRepository
}

type Command struct {
	Query string
	Limit int
}

type Result struct {
	UserUuid uuid.UUID `json:"userUuid"`
	Avatar   *string   `json:"avatar"`
	Name     string    `json:"name"`
}

func NewHandler(repository domain.ProfileRepository) *Handler {
	return &Handler{
		repository: repository,
	}
}

func (handler *Handler) Handle(ctx context.Context, command Command) ([]Result, error) {
	query := strings.TrimSpace(command.Query)

	validator := validation.NewValidator(map[string]any{
		"query": query,
		"limit": command.Limit,
	}, map[string][]rules.Rule{
		"query": {rules.Required(), rules.Min(1), rules.Max(64), rules.Regex(rules.SlugWithSpacesPattern)},
		"limit": {rules.Required(), rules.Min(1), rules.Max(20)},
	})

	if err := validator.Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return nil, validationError
		}

		return nil, fmt.Errorf("validating search profiles command: %w", err)
	}

	profiles, err := handler.repository.Search(ctx, query, command.Limit)
	if err != nil {
		return nil, fmt.Errorf("searching profiles by query %q: %w", query, err)
	}

	results := make([]Result, 0, len(profiles))
	for _, profile := range profiles {
		results = append(results, Result{
			UserUuid: profile.UserUuid,
			Avatar:   profile.Avatar,
			Name:     profile.Name,
		})
	}

	return results, nil
}
