package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/EugeneNail/vox/auth/internal/application/usecases/refresh"
	"github.com/EugeneNail/vox/lib-common/validation"
)

type refreshPayload struct {
	RefreshToken string `json:"refreshToken"`
}

type RefreshHandler struct {
	usecase *refresh.Handler
}

func NewRefreshHandler(usecase *refresh.Handler) *RefreshHandler {
	return &RefreshHandler{
		usecase: usecase,
	}
}

// Refresh decodes the request and calls the use-case.
func (handler *RefreshHandler) Handle(request *http.Request) (int, any) {
	var payload refreshPayload
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		return http.StatusBadRequest, fmt.Errorf("decoding payload: %w", err)
	}

	loginToken, err := handler.usecase.Handle(request.Context(), refresh.Query{
		RefreshToken: payload.RefreshToken,
	})
	if err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		if errors.Is(err, refresh.ErrInvalidToken) {
			validationError = validation.NewError()
			validationError.AddViolation("refreshToken", "Invalid refresh token")
			return http.StatusUnauthorized, validationError.Violations()
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the Refresh usecase: %w", err)
	}

	return http.StatusOK, map[string]string{"loginToken": loginToken}
}
