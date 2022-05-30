package ingestionHandler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/projectkeas/ingestion/sdk"
	"github.com/projectkeas/ingestion/services/eventPublisher"
	"github.com/projectkeas/ingestion/services/eventTypes"
	"github.com/projectkeas/ingestion/services/ingestionPolicies"
	"github.com/projectkeas/ingestion/services/metadataVerification"
	log "github.com/projectkeas/sdks-service/logger"
	"github.com/projectkeas/sdks-service/server"
	jsonSchema "github.com/santhosh-tekuri/jsonschema/v5"
	"go.uber.org/zap"
)

var Validator = validator.New()

func New(server *server.Server) func(context *fiber.Ctx) error {

	// Setup the services, starting with metadata validation
	metadataValidation := metadataVerification.New()

	// payload validation setup
	eventTypesService, err := server.GetService(eventTypes.SERVICE_NAME)
	if err != nil {
		panic(err)
	}
	eventValidation := (*eventTypesService).(eventTypes.EventTypeService)

	// ingestion engine setup
	ingestionService, err := server.GetService(ingestionPolicies.SERVICE_NAME)
	if err != nil {
		panic(err)
	}
	ingestionPolicyEngine := (*ingestionService).(ingestionPolicies.IngestionPolicyService)

	// publisher setup
	nc, err := server.GetService(eventPublisher.SERVICE_NAME)
	if err != nil {
		panic(err)
	}
	client := (*nc).(eventPublisher.EventPublisherService)

	return func(context *fiber.Ctx) error {
		context.Accepts("application/json")
		errorResult := map[string]interface{}{
			"message": "An error occurred whilst processing your request",
		}

		// Ensure that we can parse the event
		event := new(sdk.EventEnvelope)
		requestBody := map[string]interface{}{}
		err := context.BodyParser(&requestBody)
		if err != nil {
			log.Logger.Error("Unable to parse request body", zap.Error(err))
			errorResult["reason"] = "request-body"
			return context.Status(fiber.StatusBadRequest).JSON(errorResult)
		}

		// Validate that both sections are there in the request body
		meta, found := requestBody["metadata"]
		if !found {
			errorResult["message"] = "The metadata section is missing from the request body"
			errorResult["reason"] = "metadata-missing"
			return context.Status(fiber.StatusBadRequest).JSON(errorResult)
		}

		payload, found := requestBody["payload"]
		if !found {
			errorResult["message"] = "The payload section is missing from the request body"
			errorResult["reason"] = "payload-missing"
			return context.Status(fiber.StatusBadRequest).JSON(errorResult)
		}
		mapstructure.Decode(payload, &event.Payload)
		mapstructure.Decode(meta, &event.Metadata)

		// Validate the metadata section of the payload
		err = metadataValidation.Validate(meta)
		if err != nil {
			validationError, castSuccess := err.(*jsonSchema.ValidationError)
			if castSuccess {
				errorResult["message"] = "The specified payload does not match event schema"
				errorResult["reason"] = "metadata"
				errorResult["errors"] = validationError.Causes
				return context.Status(fiber.StatusBadRequest).JSON(errorResult)
			} else {
				log.Logger.Error("Unable to validate schema", zap.Error(err))
				errorResult["reason"] = "metadata-failure"
				return context.Status(fiber.StatusBadRequest).JSON(errorResult)
			}
		}

		// Validate that the request matches the defined schema
		err = eventValidation.Validate(*event)
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
		ingestionDecision, err := ingestionPolicyEngine.GetDecision(*event)
		if err != nil {
			log.Logger.Error("Unable to make ingestion decision", zap.Error(err))
			errorResult["message"] = "Unable to make ingestion decision"
			errorResult["reason"] = "ingestion-service-failure"
			return context.Status(fiber.StatusInternalServerError).JSON(errorResult)
		}

		// Forward the event through to the NATS cluster
		if ingestionDecision.Allow {
			if !client.Publish(event.Metadata, context.Body()) {
				errorResult["reason"] = "publish"
				return context.Status(fiber.StatusInternalServerError).JSON(errorResult)
			}

			context.Status(fiber.StatusAccepted)
			return nil
		}

		errorResult["message"] = "The event was rejected by an ingestion policy"
		errorResult["reason"] = "ingestion-service-rejected"
		return context.Status(fiber.StatusBadRequest).JSON(errorResult)
	}
}
