package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/EugeneNail/vox/auth/internal/application/usecases/authenticate"
	"github.com/EugeneNail/vox/lib-common/validation"
)

type authenticatePayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthenticateHandler struct {
	usecase *authenticate.Handler
}

func NewAuthenticateHandler(usecase *authenticate.Handler) *AuthenticateHandler {
	return &AuthenticateHandler{
		usecase: usecase,
	}
}

// Authenticate decodes the request and calls the use-case.
func (handler *AuthenticateHandler) Handle(request *http.Request) (int, any) {
	var payload authenticatePayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	loginToken, refreshToken, err := handler.usecase.Handle(request.Context(), authenticate.Query{
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
