package upload_image

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/EugeneNail/vox/attachments/internal/domain"
	"github.com/EugeneNail/vox/lib-common/validation"
	"github.com/samborkent/uuidv7"
)

const maxImageFileSize = 10 * 1024 * 1024

// Handler stores uploaded files on disk.
type Handler struct {
	filesDir string
}

// Command contains the input required to upload a file.
type Command struct {
	ContentType string
	Data        []byte
}

// NewHandler constructs an upload_image handler with its dependencies.
func NewHandler(filesDir string) *Handler {
	return &Handler{filesDir: filesDir}
}

// Handle validates the upload and stores the file.
func (handler *Handler) Handle(command Command) (string, error) {
	validationError := validation.NewError()

	extension, ok := domain.ImageExtensionFromContentType(command.ContentType)
	if !ok {
		validationError.AddViolation("contentType", "must be a supported image type")
	}

	if len(command.Data) > maxImageFileSize {
		validationError.AddViolation("size", fmt.Sprintf("must not exceed %d bytes", maxImageFileSize))
	}

	if len(validationError.Violations()) > 0 {
		return "", validationError
	}

	storageName := uuidv7.New().String() + extension.String()
	fullPath := filepath.Join(handler.filesDir, storageName)
	if err := os.WriteFile(fullPath, command.Data, 0o644); err != nil {
		return "", fmt.Errorf("writing uploaded file %q: %w", storageName, err)
	}

	return storageName, nil
}
