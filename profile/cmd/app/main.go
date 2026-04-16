package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	lib_middleware "github.com/EugeneNail/vox/lib-common/http/middleware"
	"github.com/EugeneNail/vox/profile/internal/application/usecases/create_profile"
	"github.com/EugeneNail/vox/profile/internal/application/usecases/edit_profile"
	"github.com/EugeneNail/vox/profile/internal/application/usecases/get_profiles"
	"github.com/EugeneNail/vox/profile/internal/application/usecases/search_profiles"
	"github.com/EugeneNail/vox/profile/internal/domain/events"
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
	editProfileHandler := edit_profile.NewHandler(profileRepository)
	getProfilesHandler := get_profiles.NewHandler(profileRepository)
	searchProfilesHandler := search_profiles.NewHandler(profileRepository)

	// --- Section: Event consumers ---
	userCreatedConsumer := redis_infrastructure.NewUserCreatedConsumer(redisClient, func(ctx context.Context, event events.UserCreated) error {
		return createProfileHandler.Handle(ctx, create_profile.Command{
			UserUuid: event.UserUuid,
			Name:     event.Name,
		})
	})
	userCreatedConsumer.ListenAndConsume(context.Background())

	// --- Section: HTTP transport ---
	editProfileHttpHandler := transport_http.NewEditProfileHandler(editProfileHandler)
	getProfilesHttpHandler := transport_http.NewGetProfilesHandler(getProfilesHandler)
	searchProfilesHttpHandler := transport_http.NewSearchProfilesHandler(searchProfilesHandler)

	// --- Section: HTTP routes ---
	webServer := http.NewServeMux()
	webServer.HandleFunc("POST   /api/v1/profile/profiles/batch", lib_middleware.WriteJsonResponse(getProfilesHttpHandler.Handle))
	webServer.HandleFunc("GET    /api/v1/profile/search", lib_middleware.WriteJsonResponse(searchProfilesHttpHandler.Handle))
	webServer.Handle("PUT    /api/v1/profile/profiles/me", lib_middleware.RequireAuthenticatedUser(lib_middleware.WriteJsonResponse(editProfileHttpHandler.Handle)))

	// --- Section: HTTP server ---
	address := fmt.Sprintf("0.0.0.0:%d", configuration.App.Port)

	log.Printf("profile service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}
