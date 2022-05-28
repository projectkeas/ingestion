package ingestionHandler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/projectkeas/ingestion/sdk"
	"github.com/projectkeas/ingestion/services/ingestionPolicies"
	log "github.com/projectkeas/sdks-service/logger"
	"github.com/projectkeas/sdks-service/server"
	"go.uber.org/zap"
)

type ValidationError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

type ErrorResult struct {
	Message          string            `json:"message"`
	ValidationErrors []ValidationError `json:"validationErrors,omitempty"`
}

var Validator = validator.New()

func New(server *server.Server) func(context *fiber.Ctx) error {
	return func(context *fiber.Ctx) error {
		context.Accepts("application/json")
		errorResult := &ErrorResult{
			Message: "An error occurred whilst processing your request",
		}

		// Ensure that we can parse the body
		body := new(sdk.EventEnvelope)
		err := context.BodyParser(&body)
		if err != nil {
			errorResult.Message = err.Error()
			return context.Status(fiber.StatusBadRequest).JSON(errorResult)
		}

		service, err := server.GetService(ingestionPolicies.SERVICE_NAME)
		if err != nil {
			log.Logger.Error("Unable to get the service from the request context", zap.Error(err))
			return context.Status(fiber.StatusBadRequest).JSON(errorResult)
		}

		// TODO :: validation of request payload

		ingestionPolicyEngine := service.(ingestionPolicies.IngestionPolicyService)
		ingestionDecision, err := ingestionPolicyEngine.GetDecision(*body)

		if err != nil {
			log.Logger.Error("Unable to make ingestion decision", zap.Error(err))
			return context.Status(fiber.StatusBadRequest).JSON(errorResult)
		}

		if ingestionDecision.Allow {
			// TODO :: Send through to our NATS cluster for further processing
			context.Status(204)
			return nil
		}

		context.Response().Header.Add("X-Rejected-Ingestion", "true")
		return context.Status(400).JSON(map[string]string{
			"reason": "prevented by ingestion policy",
		})
	}
}
