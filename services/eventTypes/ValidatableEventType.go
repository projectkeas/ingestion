package eventTypes

import (
	"github.com/projectkeas/ingestion/sdk"
)

type validatableEventType struct {
	SubTypes []string
	Sources  []string
	Schema   string
}

func (vt validatableEventType) Validate(event sdk.EventEnvelope) bool {

	if len(vt.SubTypes) > 0 {
		if !contains(vt.SubTypes, event.Metadata.EventSubType) {
			return false
		}
	}

	if len(vt.Sources) > 0 {
		if !contains(vt.Sources, event.Metadata.Source) {
			return false
		}
	}

	return true
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
