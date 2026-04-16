package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/EugeneNail/vox/lib-common/authentication"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/profile/internal/application/usecases/edit_profile"
)

type EditProfileHandler struct {
	usecase *edit_profile.Handler
}

type editProfilePayload struct {
	Name     string  `json:"name"`
	Nickname string  `json:"nickname"`
	Avatar   *string `json:"avatar"`
}

func NewEditProfileHandler(usecase *edit_profile.Handler) *EditProfileHandler {
	return &EditProfileHandler{
		usecase: usecase,
	}
}

func (handler *EditProfileHandler) Handle(request *http.Request) (int, any) {
	var payload editProfilePayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	userUuid, ok := authentication.UserUuidFromContext(request.Context())
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("extracting authenticated user uuid from request context")
	}

	_, err := handler.usecase.Handle(request.Context(), edit_profile.Command{
		UserUuid: userUuid,
		Name:     payload.Name,
		Nickname: payload.Nickname,
		Avatar:   payload.Avatar,
	})
	if err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		if errors.Is(err, edit_profile.ErrProfileNotFound) {
			return http.StatusNotFound, fmt.Errorf("profile for user %q not found", userUuid)
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the EditProfile usecase for user %q: %w", userUuid, err)
	}

	return http.StatusNoContent, nil
}
