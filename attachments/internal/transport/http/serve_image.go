package http

import (
	"net/http"

	"github.com/EugeneNail/vox/attachments/internal/application/usecases/serve_image"
)

// ServeImageHandler serves stored files.
type ServeImageHandler struct {
	usecase *serve_image.Handler
}

// NewServeImageHandler constructs a serve_image handler with its dependencies.
func NewServeImageHandler(usecase *serve_image.Handler) *ServeImageHandler {
	return &ServeImageHandler{usecase: usecase}
}

// Handle resolves the file path and streams the file to the response.
func (handler *ServeImageHandler) Handle(writer http.ResponseWriter, request *http.Request) {
	filePath, err := handler.usecase.Handle(request.Context(), serve_image.Command{
		Name: request.PathValue("name"),
	})
	if err != nil {
		http.Error(writer, "not found", http.StatusNotFound)
		return
	}

	http.ServeFile(writer, request, filePath)
}
