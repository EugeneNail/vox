package http

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/EugeneNail/vox/attachments/internal/application/usecases/upload_image"
)

// UploadImageHandler stores uploaded files.
type UploadImageHandler struct {
	usecase *upload_image.Handler
}

// NewUploadImageHandler constructs an upload_image handler with its dependencies.
func NewUploadImageHandler(usecase *upload_image.Handler) *UploadImageHandler {
	return &UploadImageHandler{usecase: usecase}
}

// Handle validates the multipart request and calls the use-case.
func (handler *UploadImageHandler) Handle(request *http.Request) (int, any) {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" {
		return http.StatusUnsupportedMediaType, fmt.Errorf("content type must be multipart/form-data")
	}

	if err := request.ParseMultipartForm(16 * 1024 * 1024); err != nil {
		return http.StatusBadRequest, fmt.Errorf("parsing multipart form: %w", err)
	}

	file, header, err := request.FormFile("file")
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("reading multipart file field %q: %w", "file", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("reading uploaded file: %w", err)
	}

	contentType := strings.TrimSpace(header.Header.Get("Content-Type"))
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = http.DetectContentType(data)
	}

	name, err := handler.usecase.Handle(upload_image.Command{
		ContentType: contentType,
		Data:        data,
	})
	if err != nil {
		return http.StatusUnprocessableEntity, err
	}

	return http.StatusCreated, name
}
