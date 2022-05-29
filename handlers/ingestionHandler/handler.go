package ingestionHandler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/projectkeas/ingestion/sdk"
	"github.com/projectkeas/ingestion/services/eventTypes"
	"github.com/projectkeas/ingestion/services/ingestionPolicies"
	log "github.com/projectkeas/sdks-service/logger"
	"github.com/projectkeas/sdks-service/server"
	jsonSchema "github.com/santhosh-tekuri/jsonschema/v5"
	"go.uber.org/zap"
)

var Validator = validator.New()

func New(server *server.Server) func(context *fiber.Ctx) error {
	return func(context *fiber.Ctx) error {
		context.Accepts("application/json")
		errorResult := map[string]interface{}{
			"message": "An error occurred whilst processing your request",
		}

		// Ensure that we can parse the event
		event := new(sdk.EventEnvelope)
		err := context.BodyParser(&event)
		if err != nil {
			log.Logger.Error("Unable to parse request body", zap.Error(err))
			errorResult["reason"] = "request-body"
			return context.Status(fiber.StatusBadRequest).JSON(errorResult)
		}

		// Validate that the request matches the defined schema
		eventTypesService, err := server.GetService(eventTypes.SERVICE_NAME)
		if err != nil {
			log.Logger.Error("Unable to get the eventTypes service from the request context", zap.Error(err))
			errorResult["reason"] = "event-type"
			return context.Status(fiber.StatusInternalServerError).JSON(errorResult)
		}
		eventTypesEngine := eventTypesService.(eventTypes.EventTypeService)
		err = eventTypesEngine.Validate(*event)
		if err != nil {
			validationError, castSuccess := err.(*jsonSchema.ValidationError)
			if castSuccess {
				errorResult["message"] = "The specified payload does not match event schema"
				errorResult["reason"] = "event-validation"
				errorResult["errors"] = validationError.Causes
				return context.Status(fiber.StatusBadRequest).JSON(errorResult)
			} else {
				log.Logger.Error("Unable to validate schema", zap.Error(err))
				errorResult["reason"] = "event-validation-failure"
				return context.Status(fiber.StatusBadRequest).JSON(errorResult)
			}
		}

		// Ensure that we are allowed to ingest the event
		ingestionService, err := server.GetService(ingestionPolicies.SERVICE_NAME)
		if err != nil {
			log.Logger.Error("Unable to get the ingestion service from the request context", zap.Error(err))
			errorResult["reason"] = "ingestion-service"
			return context.Status(fiber.StatusInternalServerError).JSON(errorResult)
		}
		ingestionPolicyEngine := ingestionService.(ingestionPolicies.IngestionPolicyService)
		ingestionDecision, err := ingestionPolicyEngine.GetDecision(*event)
		if err != nil {
			log.Logger.Error("Unable to make ingestion decision", zap.Error(err))
			errorResult["message"] = "Unable to make ingestion decision"
			errorResult["reason"] = "ingestion-service-failure"
			return context.Status(fiber.StatusInternalServerError).JSON(errorResult)
		}

		// Forward the event through to the NATS cluster
		if ingestionDecision.Allow {
			// TODO :: implement
			context.Status(fiber.StatusNoContent)
			return nil
		}

		errorResult["message"] = "The event was rejected by an ingestion policy"
		errorResult["reason"] = "ingestion-service-rejected"
		return context.Status(fiber.StatusBadRequest).JSON(errorResult)
	}
}
