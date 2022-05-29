package ingestionHandler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/projectkeas/ingestion/sdk"
	"github.com/projectkeas/ingestion/services/eventTypes"
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

		// Ensure that we can parse the event
		event := new(sdk.EventEnvelope)
		err := context.BodyParser(&event)
		if err != nil {
			errorResult.Message = err.Error()
			return context.Status(fiber.StatusBadRequest).JSON(errorResult)
		}

		// Validate that the request matches the defined schema
		eventTypesService, err := server.GetService(eventTypes.SERVICE_NAME)
		if err != nil {
			log.Logger.Error("Unable to get the eventTypes service from the request context", zap.Error(err))
			return context.Status(fiber.StatusInternalServerError).JSON(errorResult)
		}
		eventTypesEngine := eventTypesService.(eventTypes.EventTypeService)
		if !eventTypesEngine.Validate(*event) {
			errorResult.Message = "The event supplied does not match any known schema"
			return context.Status(fiber.StatusBadRequest).JSON(errorResult)
		}

		// Ensure that we are allowed to ingest the event
		ingestionService, err := server.GetService(ingestionPolicies.SERVICE_NAME)
		if err != nil {
			log.Logger.Error("Unable to get the ingestion service from the request context", zap.Error(err))
			return context.Status(fiber.StatusInternalServerError).JSON(errorResult)
		}
		ingestionPolicyEngine := ingestionService.(ingestionPolicies.IngestionPolicyService)
		ingestionDecision, err := ingestionPolicyEngine.GetDecision(*event)
		if err != nil {
			log.Logger.Error("Unable to make ingestion decision", zap.Error(err))
			errorResult.Message = "Unable to make ingestion decision"
			return context.Status(fiber.StatusInternalServerError).JSON(errorResult)
		}

		// Forward the event through to the NATS cluster
		if ingestionDecision.Allow {
			// TODO :: implement
			context.Status(fiber.StatusNoContent)
			return nil
		}

		context.Response().Header.Add("X-Rejected-Ingestion", "true")
		return context.Status(fiber.StatusBadRequest).JSON(map[string]string{
			"reason": "The event was rejected by an ingestion policy",
		})
	}
}
