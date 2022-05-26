package events

import (
	"encoding/json"
	"fmt"
)

const (
	IncidentEventTypes_Created string = "Created"
	IncidentEventTypes_Deleted string = "Deleted"
	IncidentEventTypes_Updated string = "Updated"
)

type IncidentCreated struct {
	Name string
}

type IncidentDeleted struct {
	Name string
}

type IncidentUpdated struct {
	Name string
}

func NewIncidentEventFromType(eventType string, payload json.RawMessage) (interface{}, error) {
	switch eventType {
	case IncidentEventTypes_Created:
		return IncidentCreated{Name: IncidentEventTypes_Created}, nil
	case IncidentEventTypes_Deleted:
		return IncidentDeleted{Name: IncidentEventTypes_Deleted}, nil
	case IncidentEventTypes_Updated:
		return IncidentUpdated{Name: IncidentEventTypes_Updated}, nil
	}
	return nil, fmt.Errorf("cannot parse event type '%s' for type 'Incident'", eventType)
}
