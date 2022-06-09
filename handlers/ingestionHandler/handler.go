package ingestionHandler

import (
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	spec "github.com/cloudevents/sdk-go/v2/binding/spec"
	cee "github.com/cloudevents/sdk-go/v2/event"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/projectkeas/ingestion/services/eventPublisher"
	"github.com/projectkeas/ingestion/services/eventTypes"
	"github.com/projectkeas/ingestion/services/ingestionPolicies"
	log "github.com/projectkeas/sdks-service/logger"
	"github.com/projectkeas/sdks-service/server"
	jsonSchema "github.com/santhosh-tekuri/jsonschema/v5"
	"go.uber.org/zap"
)

var Validator = validator.New()

func New(server *server.Server) func(context *fiber.Ctx) error {

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

	// cloudevents setup
	var specs = spec.New().Version("1.0")

	dataContentTypeHeader := "ce-" + specs.AttributeFromKind(spec.DataContentType).Name()
	dataSchemaHeader := "ce-" + specs.AttributeFromKind(spec.DataSchema).Name()
	idHeader := "ce-" + specs.AttributeFromKind(spec.ID).Name()
	sourceHeader := "ce-" + specs.AttributeFromKind(spec.Source).Name()
	specVersionHeader := "ce-" + specs.AttributeFromKind(spec.SpecVersion).Name()
	subjectHeader := "ce-" + specs.AttributeFromKind(spec.Subject).Name()
	typeHeader := "ce-" + specs.AttributeFromKind(spec.Type).Name()
	timeHeader := "ce-" + specs.AttributeFromKind(spec.Time).Name()

	return func(context *fiber.Ctx) error {
		context.Accepts("application/json")
		errorResult := map[string]interface{}{
			"message": "An error occurred whilst processing your request",
		}

		// Parse the request body
		requestBody := map[string]interface{}{}
		err := context.BodyParser(&requestBody)
		if err != nil {
			log.Logger.Error("Unable to parse request body", zap.Error(err))
			errorResult["reason"] = "request-body"
			return context.Status(fiber.StatusBadRequest).JSON(errorResult)
		}

		cloudEvent := cloudevents.NewEvent()
		cloudEvent.SetData(*cloudevents.StringOfApplicationJSON(), requestBody)

		// Map all the headers to the cloud event
		for key, value := range context.GetReqHeaders() {
			if strings.EqualFold(key, dataContentTypeHeader) {
				cloudEvent.SetDataContentType(value)
			} else if strings.EqualFold(key, dataSchemaHeader) {
				cloudEvent.SetDataSchema(value)
			} else if strings.EqualFold(key, idHeader) {
				cloudEvent.SetID(value)
			} else if strings.EqualFold(key, sourceHeader) {
				cloudEvent.SetSource(value)
			} else if strings.EqualFold(key, specVersionHeader) {
				cloudEvent.SetSpecVersion(value)
			} else if strings.EqualFold(key, subjectHeader) {
				cloudEvent.SetSubject(value)
			} else if strings.EqualFold(key, typeHeader) {
				cloudEvent.SetType(value)
			} else if strings.EqualFold(key, timeHeader) {
				t, err := time.Parse(time.RFC3339, value)
				if err == nil {
					cloudEvent.SetTime(t.UTC())
				} else {
					cloudEvent.SetTime(time.Now().UTC())
				}
			}
		}

		// Validate the cloud event has enough information
		err = cloudEvent.Validate()
		if err != nil {
			// TODO :: see if there is a nice way of parsing this
			errorResult["message"] = "The request does not conform to a valid cloudevent"
			errorResult["reason"] = "cloud-event-validation"
			ve := err.(cee.ValidationError)
			errors := []map[string]string{}
			for key, value := range ve {
				errors = append(errors, map[string]string{
					"attribute": key,
					"error":     value.Error(),
				})
			}
			errorResult["errors"] = errors
			return context.Status(fiber.StatusBadRequest).JSON(errorResult)
		}

		// Validate that the request matches the defined schema
		err = eventValidation.Validate(cloudEvent, requestBody)
		if err != nil {
			// TODO :: have a global option for allowing unregistered event types
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
		ingestionDecision, err := ingestionPolicyEngine.GetDecision(cloudEvent, requestBody)
		if err != nil {
			log.Logger.Error("Unable to make ingestion decision", zap.Error(err))
			errorResult["message"] = "Unable to make ingestion decision"
			errorResult["reason"] = "ingestion-service-failure"
			return context.Status(fiber.StatusInternalServerError).JSON(errorResult)
		}

		// Forward the event through to the NATS cluster
		if ingestionDecision.Allow {
			if !client.Publish(cloudEvent) {
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
