package authenticationHandler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	log "github.com/projectkeas/sdks-service/logger"
	"github.com/projectkeas/sdks-service/server"
)

func New(server *server.Server) func(context *fiber.Ctx) error {

	return func(context *fiber.Ctx) error {

		token := server.GetConfiguration().GetStringValueOrDefault("ingestion.auth.token", "")
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
