package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/EugeneNail/vox/auth/internal/application/usecases/create_user"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/validation"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/validation/rules"
	"net/http"
)

type createUserPayload struct {
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"passwordConfirmation"`
}

// CreateUser decodes the request, applies transport validation, and calls the use-case.
func (handler *Handler) CreateUser(request *http.Request) (int, any) {
	var payload createUserPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	validator := validation.NewValidator(map[string]any{
		"email":                payload.Email,
		"password":             payload.Password,
		"passwordConfirmation": payload.PasswordConfirmation,
	}, map[string][]rules.Rule{
		"email":                {rules.Required(), rules.Max(256)},
		"password":             {rules.Required(), rules.Max(256)},
		"passwordConfirmation": {rules.Required(), rules.Max(256)},
	})

	if err := validator.Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusBadRequest, validationError.Violations()
		}

		return http.StatusInternalServerError, fmt.Errorf("validating create user payload: %w", err)
	}

	userUuid, err := handler.createUserHandler.Handle(request.Context(), create_user.Command{
		Email:                payload.Email,
		Password:             payload.Password,
		PasswordConfirmation: payload.PasswordConfirmation,
	})
	if err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		if errors.Is(err, create_user.ErrEmailAlreadyExists) {
			validationError = validation.NewError()
			validationError.AddViolation("email", "The email is already taken")
			return http.StatusConflict, validationError.Violations()
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the CreateUser usecase: %w", err)
	}

	return http.StatusCreated, userUuid
}
