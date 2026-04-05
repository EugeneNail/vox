package main

import (
	"fmt"
	"github.com/EugeneNail/vox/auth/internal/application/usecases/authenticate"
	"github.com/EugeneNail/vox/auth/internal/application/usecases/create_user"
	"github.com/EugeneNail/vox/auth/internal/application/usecases/refresh"
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
	authenticateHandler := authenticate.NewHandler(userRepository)
	refreshHandler := refresh.NewHandler(userRepository)
	httpHandler := transport_http.NewHandler(createUserHandler, authenticateHandler, refreshHandler)

	webServer := http.NewServeMux()
	webServer.HandleFunc(
		"POST /api/v1/auth/users",
		middleware.RejectLargeRequest(2048, middleware.WriteJsonResponse(httpHandler.CreateUser)),
	)
	webServer.HandleFunc(
		"POST /api/v1/auth/users/authenticate",
		middleware.RejectLargeRequest(2048, middleware.WriteJsonResponse(httpHandler.Authenticate)),
	)
	webServer.HandleFunc(
		"POST /api/v1/auth/refresh",
		middleware.RejectLargeRequest(4096, middleware.WriteJsonResponse(httpHandler.Refresh)),
	)

	address := fmt.Sprintf("0.0.0.0:%d", configuration.App.Port)

	log.Printf("auth service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}
