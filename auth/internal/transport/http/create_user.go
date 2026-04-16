package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/EugeneNail/vox/auth/internal/application/usecases/create_user"
	"github.com/EugeneNail/vox/lib-common/validation"
)

type createUserPayload struct {
	Name                 string `json:"name"`
	Nickname             string `json:"nickname"`
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"passwordConfirmation"`
}

type CreateUserHandler struct {
	usecase *create_user.Handler
}

func NewCreateUserHandler(usecase *create_user.Handler) *CreateUserHandler {
	return &CreateUserHandler{
		usecase: usecase,
	}
}

// CreateUser decodes the request and calls the use-case.
func (handler *CreateUserHandler) Handle(request *http.Request) (int, any) {
	var payload createUserPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	userUuid, err := handler.usecase.Handle(request.Context(), create_user.Command{
		Name:                 payload.Name,
		Nickname:             payload.Nickname,
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
