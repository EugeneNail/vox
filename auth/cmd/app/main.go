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
	webServer.HandleFunc("/auth/ping", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("pong"))
	})

	address := ":" + strconv.Itoa(configuration.App.Port)

	log.Printf("auth service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}
