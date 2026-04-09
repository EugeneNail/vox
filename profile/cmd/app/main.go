package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/EugeneNail/vox/profile/internal/application/usecases/create_profile"
	"github.com/EugeneNail/vox/profile/internal/domain"
	"github.com/EugeneNail/vox/profile/internal/infrastructure/config"
	"github.com/EugeneNail/vox/profile/internal/infrastructure/postgres"
	redis_infrastructure "github.com/EugeneNail/vox/profile/internal/infrastructure/redis"
	transport_http "github.com/EugeneNail/vox/profile/internal/transport/http"
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
	profileRepository := postgres.NewProfileRepository(database)

	// --- Section: Application use-cases ---
	createProfileHandler := create_profile.NewHandler(profileRepository)

	// --- Section: Event consumers ---
	userCreatedConsumer := redis_infrastructure.NewUserCreatedConsumer(redisClient, func(ctx context.Context, event domain.UserCreatedEvent) error {
		return createProfileHandler.Handle(ctx, create_profile.Command{
			UserUuid: event.UserUuid,
		})
	})
	userCreatedConsumer.ListenAndConsume(context.Background())

	// --- Section: HTTP transport ---
	httpHandler := transport_http.NewHandler(profileRepository)
	_ = httpHandler

	// --- Section: HTTP routes ---
	webServer := http.NewServeMux()

	// --- Section: HTTP server ---
	address := fmt.Sprintf("0.0.0.0:%d", configuration.App.Port)

	log.Printf("profile service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}
