package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/EugeneNail/vox/auth/internal/application/usecases/authenticate"
	"github.com/EugeneNail/vox/auth/internal/application/usecases/create_user"
	"github.com/EugeneNail/vox/auth/internal/application/usecases/refresh"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/config"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/postgres"
	redis_infrastructure "github.com/EugeneNail/vox/auth/internal/infrastructure/redis"
	transport_http "github.com/EugeneNail/vox/auth/internal/transport/http"
	"github.com/EugeneNail/vox/lib-common/http/middleware"
)

func main() {
	// --- Section: Runtime configuration ---
	configuration, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	// --- Section: External clients ---
	database, err := postgres.NewDatabase(configuration.Postgres)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	redisClient, err := redis_infrastructure.NewClient(configuration.Redis)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()

	// --- Section: Repositories ---
	userRepository := postgres.NewUserRepository(database)

	// --- Section: Event delivery ---
	userCreatedPublisher := redis_infrastructure.NewUserCreatedPublisher(redisClient)

	// --- Section: Application use-cases ---
	createUserHandler := create_user.NewHandler(userRepository, userCreatedPublisher)
	authenticateHandler := authenticate.NewHandler(userRepository)
	refreshHandler := refresh.NewHandler(userRepository)

	// --- Section: HTTP transport ---
	createUserHttpHandler := transport_http.NewCreateUserHandler(createUserHandler)
	authenticateHttpHandler := transport_http.NewAuthenticateHandler(authenticateHandler)
	refreshHttpHandler := transport_http.NewRefreshHandler(refreshHandler)

	// --- Section: HTTP routes ---
	webServer := http.NewServeMux()
	webServer.HandleFunc("POST   /api/v1/auth/users", middleware.RejectLargeRequest(2048, middleware.WriteJsonResponse(createUserHttpHandler.Handle)))
	webServer.HandleFunc("POST   /api/v1/auth/users/authenticate", middleware.RejectLargeRequest(2048, middleware.WriteJsonResponse(authenticateHttpHandler.Handle)))
	webServer.HandleFunc("POST   /api/v1/auth/refresh", middleware.RejectLargeRequest(4096, middleware.WriteJsonResponse(refreshHttpHandler.Handle)))

	// --- Section: HTTP server ---
	address := fmt.Sprintf("0.0.0.0:%d", configuration.App.Port)

	log.Printf("auth service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}
