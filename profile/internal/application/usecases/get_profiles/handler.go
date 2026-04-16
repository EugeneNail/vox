package get_profiles

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/profile/internal/domain"
	"github.com/google/uuid"
)

type Handler struct {
	repository domain.ProfileRepository
}

type Query struct {
	UserUuids []uuid.UUID
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

func (handler *Handler) Handle(ctx context.Context, query Query) ([]Result, error) {
	validationError := validation.NewError()

	if len(query.UserUuids) == 0 {
		validationError.AddViolation("userUuids", "Must contain at least one user UUID")
	}

	if len(query.UserUuids) > 500 {
		validationError.AddViolation("userUuids", "Must contain at most 500 user UUIDs")
	}

	seenUserUuids := make(map[uuid.UUID]struct{}, len(query.UserUuids))
	for index, userUuid := range query.UserUuids {
		fieldName := fmt.Sprintf("userUuids.%d", index)

		if userUuid == uuid.Nil {
			validationError.AddViolation(fieldName, "Must not be empty")
			continue
		}

		if _, ok := seenUserUuids[userUuid]; ok {
			validationError.AddViolation(fieldName, "Must not contain duplicates")
			continue
		}

		seenUserUuids[userUuid] = struct{}{}
	}

	if len(validationError.Violations()) > 0 {
		return nil, validationError
	}

	profiles, err := handler.repository.FindAllByUserUuids(ctx, query.UserUuids)
	if err != nil {
		return nil, fmt.Errorf("finding profiles by user uuids: %w", err)
	}

	results := make([]Result, 0, len(profiles))
	for _, userUuid := range query.UserUuids {
		profileIndex := slices.IndexFunc(profiles, func(profile domain.Profile) bool {
			return profile.UserUuid == userUuid
		})
		if profileIndex == -1 {
			continue
		}

		profile := profiles[profileIndex]
		results = append(results, Result{
			UserUuid: profile.UserUuid,
			Avatar:   profile.Avatar,
			Name:     profile.Name,
		})
	}

	return results, nil
}

func AsValidationError(err error) (validation.Error, bool) {
	var validationError validation.Error
	if errors.As(err, &validationError) {
		return validationError, true
	}

	return validation.Error{}, false
}
