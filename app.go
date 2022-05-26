package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/projectkeas/sdks-service/configuration"
	"github.com/projectkeas/sdks-service/server"

	"github.com/projectkeas/ingestion/handlers/ingestionHandler"
)

func main() {
	app := server.New("ingestion")

	app.WithEnvironmentVariableConfiguration("KEAS_")

	//app.WithConfigMap("config-1")
	//app.WithSecret("secret-1")

	app.ConfigureHandlers(func(f *fiber.App, configurationAccessor func() *configuration.ConfigurationRoot) {
		f.Post("/ingest", ingestionHandler.New(configurationAccessor))
	})

	app.Run()
}
