package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/EugeneNail/vox/attachments/internal/application/usecases/serve_image"
	"github.com/EugeneNail/vox/attachments/internal/application/usecases/upload_image"
	"github.com/EugeneNail/vox/attachments/internal/infrastructure/config"
	transport_http "github.com/EugeneNail/vox/attachments/internal/transport/http"
	"github.com/EugeneNail/vox/lib-common/http/middleware"
)

func main() {
	configuration, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	if err := configuration.Images.CreateDirectory(); err != nil {
		log.Fatal(err)
	}

	// --- Section: Application use-cases ---
	serveImageUsecase := serve_image.NewHandler(configuration.Images.Directory)
	uploadImageUsecase := upload_image.NewHandler(configuration.Images.Directory)

	// --- Section: HTTP transport ---
	serveImageHttpHandler := transport_http.NewServeImageHandler(serveImageUsecase)
	uploadImageHttpHandler := transport_http.NewUploadImageHandler(uploadImageUsecase)

	// --- Section: HTTP routes ---
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/attachments/images/{name}", serveImageHttpHandler.Handle)
	mux.HandleFunc("POST /api/v1/attachments/images", middleware.RejectLargeRequest(10*1024*1024, middleware.WriteJsonResponse(uploadImageHttpHandler.Handle)))

	// --- Section: HTTP server ---
	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", configuration.App.Port),
		Handler: mux,
	}

	log.Printf("attachments service listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
