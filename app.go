package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/projectkeas/sdks-service/server"

	"github.com/projectkeas/ingestion/handlers/authenticationHandler"
	"github.com/projectkeas/ingestion/handlers/ingestionHandler"
	"github.com/projectkeas/ingestion/services/eventPublisher"
	"github.com/projectkeas/ingestion/services/eventTypes"
	"github.com/projectkeas/ingestion/services/ingestionPolicies"
)

func main() {
	app := server.New("ingestion")

	app.WithEnvironmentVariableConfiguration("KEAS_")

	app.WithConfigMap("ingestion-cm")
	app.WithRequiredSecret("ingestion-secret")

	app.ConfigureHandlers(func(f *fiber.App, server *server.Server) {
		f.Post("/ingest", authenticationHandler.New(server), ingestionHandler.New(server))
	})

	server := app.Build()

	server.RegisterService(ingestionPolicies.SERVICE_NAME, ingestionPolicies.New())
	server.RegisterService(eventTypes.SERVICE_NAME, eventTypes.New())
	server.RegisterService(eventPublisher.SERVICE_NAME, eventPublisher.New(server.GetConfiguration()))

	server.Run()
}
