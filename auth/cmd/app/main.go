package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/EugeneNail/vox/auth/internal/infrastructure/config"
)

func main() {
	configuration, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	webServer := http.NewServeMux()

	address := ":" + strconv.Itoa(configuration.App.Port)

	log.Printf("auth service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}
