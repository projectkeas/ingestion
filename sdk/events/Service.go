package events

import (
	"encoding/json"
	"fmt"
)

const (
	ServiceEventTypes_Created string = "Created"
	ServiceEventTypes_Deleted string = "Deleted"
	ServiceEventTypes_Updated string = "Updated"
)

type ServiceCreated struct {
	Name string
}

type ServiceDeleted struct {
	Name string
}

type ServiceUpdated struct {
	Name string
}

func NewServiceEventFromType(eventType string, payload json.RawMessage) (interface{}, error) {
	switch eventType {
	case ServiceEventTypes_Created:
		return ServiceCreated{Name: ServiceEventTypes_Created}, nil
	case ServiceEventTypes_Deleted:
		return ServiceDeleted{Name: ServiceEventTypes_Deleted}, nil
	case ServiceEventTypes_Updated:
		return ServiceUpdated{Name: ServiceEventTypes_Updated}, nil
	}
	return nil, fmt.Errorf("cannot parse event type '%s' for type 'Service'", eventType)
}
