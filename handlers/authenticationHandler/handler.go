package authenticationHandler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/projectkeas/sdks-service/configuration"
	log "github.com/projectkeas/sdks-service/logger"
	"github.com/projectkeas/sdks-service/server"
)

var token string = ""

func New(server *server.Server) func(context *fiber.Ctx) error {

	config := server.GetConfiguration()
	config.RegisterChangeNotificationHandler(func(newConfig configuration.ConfigurationRoot) {
		token = newConfig.GetStringValueOrDefault("ingestion.auth.token", "")
	})

	return func(context *fiber.Ctx) error {

		if token == "" {
			log.Logger.Warn("No token has been set for authentication")
			return context.SendStatus(401)
		}

		if context.Get("Authorization") == fmt.Sprintf("ApiKey %s", token) {
			return context.Next()
		}

		return context.SendStatus(401)
	}
}
