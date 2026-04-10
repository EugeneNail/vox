package serve_image

import (
	"context"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/EugeneNail/vox/lib-common/validation/rules"
)

// Handler resolves file paths for the serve_image use-case.
type Handler struct {
	directory string
}

// Command contains the input required to serve a file.
type Command struct {
	Name string
}

// NewHandler constructs a serve_image handler with its dependencies.
func NewHandler(directory string) *Handler {
	return &Handler{directory: directory}
}

// Handle validates the file name and resolves the file path.
func (handler *Handler) Handle(_ context.Context, command Command) (string, error) {
	name := strings.TrimSpace(command.Name)
	validator := validation.NewValidator(map[string]any{
		"name": name,
	}, map[string][]rules.Rule{
		"name": {rules.Required()},
	})

	if err := validator.Validate(); err != nil {
		var validationError validation.Error
		if errors.As(err, &validationError) {
			return "", validationError
		}

		return "", fmt.Errorf("validating serve file command: %w", err)
	}

	if !isSafeFileName(name) {
		validationError := validation.NewError()
		validationError.AddViolation("name", "must be a single file name")
		return "", validationError
	}

	return filepath.Join(handler.directory, name), nil
}

func isSafeFileName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}

	if name != path.Base(name) {
		return false
	}

	return !strings.Contains(name, "\\")
}
