package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/EugeneNail/vox/auth/internal/application/authenticate"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/validation"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/validation/rules"
	"net/http"
)

type authenticatePayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Authenticate decodes the request, applies transport validation, and calls the use-case.
func (handler *Handler) Authenticate(request *http.Request) (int, any) {
	var payload authenticatePayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	validator := validation.NewValidator(map[string]any{
		"email":    payload.Email,
		"password": payload.Password,
	}, map[string][]rules.Rule{
		"email":    {rules.Required(), rules.Max(256)},
		"password": {rules.Required(), rules.Max(256)},
	})

	if err := validator.Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusBadRequest, validationError.Violations()
		}

		return http.StatusInternalServerError, fmt.Errorf("validating authenticate payload: %w", err)
	}

	loginToken, refreshToken, err := handler.authenticateHandler.Handle(request.Context(), authenticate.Query{
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		if errors.Is(err, authenticate.ErrInvalidCredentials) {
			validationError = validation.NewError()
			validationError.AddViolation("email", "Invalid credentials")
			validationError.AddViolation("password", "Invalid credentials")
			return http.StatusUnauthorized, validationError.Violations()
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the Authenticate usecase: %w", err)
	}

	return http.StatusOK, map[string]string{
		"loginToken":   loginToken,
		"refreshToken": refreshToken,
	}
}
