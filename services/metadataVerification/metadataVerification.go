package metadataVerification

import (
	jsonSchema "github.com/santhosh-tekuri/jsonschema/v5"
)

var schema string = `{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"source": {
			"type": "string",
			"pattern": "^[A-z\\-]{3,63}$"
		},
		"type": {
			"type": "string",
			"pattern": "^[A-z\\-]{3,63}$"
		},
		"subType": {
			"type": "string",
			"pattern": "^[A-z\\-]{3,63}$"
		},
		"eventTime": {
			"type": "string",
			"format": "date-time"
		},
		"eventUUID": {
			"type": "string",
			"format": "date-time"
		},
		"version": {
			"type": "string",
			"pattern": "^([0-9]{1,4}){1}\\.([0-9]{1,4}){1}\\.([0-9]{1,4}){1}$"
		}
	},
	"required": [
		"source",
		"type",
		"version"
	]
}`

type MetadataVerificationService struct {
	schema *jsonSchema.Schema
}

func New() *MetadataVerificationService {
	temp, err := jsonSchema.CompileString("schema.json", schema)
	if err != nil {
		panic(err)
	}
	return &MetadataVerificationService{
		schema: temp,
	}
}

func (mv *MetadataVerificationService) Validate(metadata interface{}) error {
	return mv.schema.Validate(metadata)
}
