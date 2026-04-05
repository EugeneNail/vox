package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/EugeneNail/vox/auth/internal/application/refresh"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/validation"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/validation/rules"
)

type refreshPayload struct {
	RefreshToken string `json:"refreshToken"`
}

// Refresh decodes the request, applies transport validation, and calls the use-case.
func (handler *Handler) Refresh(request *http.Request) (int, any) {
	var payload refreshPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	validator := validation.NewValidator(map[string]any{
		"refreshToken": payload.RefreshToken,
	}, map[string][]rules.Rule{
		"refreshToken": {rules.Required(), rules.Max(4096)},
	})

	if err := validator.Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusBadRequest, validationError.Violations()
		}

		return http.StatusInternalServerError, fmt.Errorf("validating refresh payload: %w", err)
	}

	loginToken, err := handler.refreshHandler.Handle(request.Context(), refresh.Query{
		RefreshToken: payload.RefreshToken,
	})
	if err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		if errors.Is(err, refresh.ErrInvalidRefreshToken) {
			validationError = validation.NewError()
			validationError.AddViolation("refreshToken", "Invalid refresh token")
			return http.StatusUnauthorized, validationError.Violations()
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the Refresh usecase: %w", err)
	}

	return http.StatusOK, map[string]string{"loginToken": loginToken}
}
