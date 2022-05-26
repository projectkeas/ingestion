package ingestionHandler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/projectkeas/ingestion/sdk"
	"github.com/projectkeas/sdks-service/configuration"
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

func New(configuration func() *configuration.ConfigurationRoot) func(context *fiber.Ctx) error {
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

		// Validate the body according to our rules
		err = Validator.Struct(body)
		if err != nil {
			for _, err := range err.(validator.ValidationErrors) {
				errorResult.ValidationErrors = append(errorResult.ValidationErrors, ValidationError{
					Field: err.Field(),
					Error: err.Error(),
				})
			}
			return context.Status(fiber.StatusBadRequest).JSON(errorResult)
		}

		// TODO :: Pass through OPA to see whether or not we should handle this request

		// TODO :: Send through to our NATS cluster for further processing
		context.Status(204)
		return nil
	}
}
