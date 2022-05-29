package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/projectkeas/sdks-service/server"

	"github.com/projectkeas/ingestion/handlers/ingestionHandler"
	"github.com/projectkeas/ingestion/services/eventTypes"
	"github.com/projectkeas/ingestion/services/ingestionPolicies"
)

func main() {
	app := server.New("ingestion")

	app.WithEnvironmentVariableConfiguration("KEAS_")

	//app.WithConfigMap("config-1")
	//app.WithSecret("secret-1")

	app.ConfigureHandlers(func(f *fiber.App, server *server.Server) {
		f.Post("/ingest", ingestionHandler.New(server))
	})

	server := app.Build()

	server.RegisterService(ingestionPolicies.SERVICE_NAME, ingestionPolicies.New())
	server.RegisterService(eventTypes.SERVICE_NAME, eventTypes.New())

	server.Run()
}
