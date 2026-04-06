package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/EugeneNail/vox/lib-common/http/middleware"
	"github.com/EugeneNail/vox/message/internal/application/usecases/create_message"
	"github.com/EugeneNail/vox/message/internal/infrastructure/config"
	message_middleware "github.com/EugeneNail/vox/message/internal/infrastructure/http/middleware"
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

	messageRepository := postgres.NewMessageRepository(database)
	createMessageHandler := create_message.NewHandler(messageRepository)
	httpHandler := transport_http.NewHandler(createMessageHandler)

	webServer := http.NewServeMux()
	webServer.HandleFunc(
		"GET /api/v1/message/ping",
		middleware.WriteJsonResponse(httpHandler.Ping),
	)
	webServer.HandleFunc(
		"POST /api/v1/message/chats/{chatUuid}/messages",
		message_middleware.RequireAuthenticatedUser(
			middleware.RejectLargeRequest(4096, middleware.WriteJsonResponse(httpHandler.CreateMessage)),
		),
	)

	address := fmt.Sprintf("0.0.0.0:%d", configuration.App.Port)

	log.Printf("message service listening on %s", address)
	if err := http.ListenAndServe(address, webServer); err != nil {
		log.Fatal(err)
	}
}
