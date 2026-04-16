package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/EugeneNail/vox/profile/internal/application/usecases/get_profiles"
	"github.com/EugeneNail/vox/profile/internal/transport/http/resource"
	"github.com/google/uuid"
)

type GetProfilesHandler struct {
	usecase *get_profiles.Handler
}

type getProfilesPayload struct {
	UserUuids []string `json:"userUuids"`
}

func NewGetProfilesHandler(usecase *get_profiles.Handler) *GetProfilesHandler {
	return &GetProfilesHandler{
		usecase: usecase,
	}
}

func (handler *GetProfilesHandler) Handle(request *http.Request) (int, any) {
	var payload getProfilesPayload

	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	userUuids := make([]uuid.UUID, 0, len(payload.UserUuids))

	for index, userUuidString := range payload.UserUuids {
		userUuid, err := uuid.Parse(userUuidString)
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("parsing user uuid at index %d: %w", index, err)
		}

		userUuids = append(userUuids, userUuid)
	}

	results, err := handler.usecase.Handle(request.Context(), get_profiles.Query{
		UserUuids: userUuids,
	})
	if err != nil {
		if validationError, ok := get_profiles.AsValidationError(err); ok {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the GetProfiles usecase: %w", err)
	}

	resources := make([]resource.Profile, 0, len(results))
	for _, result := range results {
		resources = append(resources, resource.Profile{
			UserUuid: result.UserUuid,
			Avatar:   result.Avatar,
			Name:     result.Name,
		})
	}

	return http.StatusOK, resources
}
