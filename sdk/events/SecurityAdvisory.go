package events

import (
	"encoding/json"
	"fmt"
)

const (
	SecurityAdvisoryEventTypes_Created string = "Created"
	SecurityAdvisoryEventTypes_Deleted string = "Deleted"
	SecurityAdvisoryEventTypes_Updated string = "Updated"
)

type SecurityAdvisoryCreated struct {
	Name string
}

type SecurityAdvisoryDeleted struct {
	Name string
}

type SecurityAdvisoryUpdated struct {
	Name string
}

func NewSecurityAdvisoryEventFromType(eventType string, payload json.RawMessage) (interface{}, error) {
	switch eventType {
	case SecurityAdvisoryEventTypes_Created:
		return SecurityAdvisoryCreated{Name: SecurityAdvisoryEventTypes_Created}, nil
	case SecurityAdvisoryEventTypes_Deleted:
		return SecurityAdvisoryDeleted{Name: SecurityAdvisoryEventTypes_Deleted}, nil
	case SecurityAdvisoryEventTypes_Updated:
		return SecurityAdvisoryUpdated{Name: SecurityAdvisoryEventTypes_Updated}, nil
	}
	return nil, fmt.Errorf("cannot parse event type '%s' for type 'SecurityAdvisory'", eventType)
}
