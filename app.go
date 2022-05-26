package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/projectkeas/sdks-service/configuration"
	"github.com/projectkeas/sdks-service/server"
)

func main() {
	app := server.New("ingestion")

	app.WithEnvironmentVariableConfiguration("KEAS_")

	//app.WithConfigMap("config-1")
	//app.WithSecret("secret-1")

	app.ConfigureHandlers(func(f *fiber.App, configurationAccessor func() *configuration.ConfigurationRoot) {
		f.Get("/", func(c *fiber.Ctx) error {
			value := configurationAccessor().GetStringValueOrDefault("log.level", "not set")
			return c.SendString(fmt.Sprintf("Hello, World 👋! Log Level is: %s", value))
		})
	})

	app.Run()
}
