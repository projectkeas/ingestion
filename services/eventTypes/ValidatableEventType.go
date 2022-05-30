package eventTypes

import (
	"fmt"

	"github.com/projectkeas/ingestion/sdk"
	jsonSchema "github.com/santhosh-tekuri/jsonschema/v5"
)

type validatableEventType struct {
	SubTypes []string
	Sources  []string
	Schema   jsonSchema.Schema
	version  string
}

func (vt validatableEventType) Validate(event sdk.EventEnvelope) error {

	if len(vt.SubTypes) > 0 {
		if !contains(vt.SubTypes, event.Metadata.SubType) {
			return fmt.Errorf("'%s' is not registered for subType '%s'", event.Metadata.Type, event.Metadata.SubType)
		}
	}

	if len(vt.Sources) > 0 {
		if !contains(vt.Sources, event.Metadata.Source) {
			return fmt.Errorf("'%s' is not registered for source '%s'", event.Metadata.Type, event.Metadata.Source)
		}
	}

	return vt.Schema.Validate(event.Payload)
}

func contains(arr []string, item string) bool {
	found := false
	for _, value := range arr {
		if value == item {
			found = true
			break
		}
	}

	return found
}
