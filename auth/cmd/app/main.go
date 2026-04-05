package main

import (
	"fmt"
	"github.com/EugeneNail/vox/auth/internal/application/create_user"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/config"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/http/middleware"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/postgres"
	transport_http "github.com/EugeneNail/vox/auth/internal/transport/http"
	"log"
	"net/http"
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

	userRepository := postgres.NewUserRepository(database)
	createUserHandler := create_user.NewHandler(userRepository)
	httpHandler := transport_http.NewHandler(createUserHandler)

	webServer := http.NewServeMux()
	webServer.HandleFunc(
		"POST /auth/users",
		middleware.RejectLargeRequest(2048, middleware.WriteJsonResponse(httpHandler.CreateUser)),
	)

	address := fmt.Sprintf("0.0.0.0:%d", configuration.App.Port)

	log.Printf("auth service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}
