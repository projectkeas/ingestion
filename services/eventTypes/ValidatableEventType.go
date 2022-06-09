package eventTypes

import (
	jsonSchema "github.com/santhosh-tekuri/jsonschema/v5"
)

type validatableEventType struct {
	schema    jsonSchema.Schema
	schemaUri string
	version   string
}

func (vt validatableEventType) Validate(data map[string]interface{}) error {
	return vt.schema.Validate(data)
}
