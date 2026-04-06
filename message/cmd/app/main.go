package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/EugeneNail/vox/message/internal/infrastructure/config"
	"github.com/EugeneNail/vox/message/internal/infrastructure/http/middleware"
	"github.com/EugeneNail/vox/message/internal/infrastructure/postgres"
	transport_http "github.com/EugeneNail/vox/message/internal/transport/http"
)

func main() {
	configuration, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	database, err := postgres.NewDatabase(configuration.Postgres)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	httpHandler := transport_http.NewHandler()

	webServer := http.NewServeMux()
	webServer.HandleFunc(
		"GET /api/v1/message/ping",
		middleware.WriteJsonResponse(httpHandler.Ping),
	)

	address := fmt.Sprintf("0.0.0.0:%d", configuration.App.Port)

	log.Printf("message service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}
