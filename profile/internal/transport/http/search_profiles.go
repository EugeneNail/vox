package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/profile/internal/application/usecases/search_profiles"
	"github.com/EugeneNail/vox/profile/internal/transport/http/resource"
)

type SearchProfilesHandler struct {
	usecase *search_profiles.Handler
}

func NewSearchProfilesHandler(usecase *search_profiles.Handler) *SearchProfilesHandler {
	return &SearchProfilesHandler{
		usecase: usecase,
	}
}

func (handler *SearchProfilesHandler) Handle(request *http.Request) (int, any) {
	rawQuery := strings.TrimSpace(request.URL.Query().Get("query"))
	rawLimit := request.URL.Query().Get("limit")

	limit := 10
	if rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("parsing limit query parameter %q: %w", rawLimit, err)
		}

		limit = parsedLimit
	}

	results, err := handler.usecase.Handle(request.Context(), search_profiles.Command{
		Query: rawQuery,
		Limit: limit,
	})
	if err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return http.StatusUnprocessableEntity, validationError.Violations()
		}

		return http.StatusInternalServerError, fmt.Errorf("handling the SearchProfiles usecase: %w", err)
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
