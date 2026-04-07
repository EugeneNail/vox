package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/EugeneNail/vox/profile/internal/infrastructure/config"
	"github.com/EugeneNail/vox/profile/internal/infrastructure/postgres"
	transport_http "github.com/EugeneNail/vox/profile/internal/transport/http"
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

	profileRepository := postgres.NewProfileRepository(database)
	httpHandler := transport_http.NewHandler(profileRepository)
	_ = httpHandler

	webServer := http.NewServeMux()

	address := fmt.Sprintf("0.0.0.0:%d", configuration.App.Port)

	log.Printf("profile service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}
