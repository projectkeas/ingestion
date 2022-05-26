package events

import (
	"encoding/json"
	"fmt"
)

const (
	DependencyEventTypes_Created string = "Created"
	DependencyEventTypes_Deleted string = "Deleted"
	DependencyEventTypes_Updated string = "Updated"
)

type DependencyCreated struct {
	Name string
}

type DependencyDeleted struct {
	Name string
}

type DependencyUpdated struct {
	Name string
}

func NewDependencyEventFromType(eventType string, payload json.RawMessage) (interface{}, error) {
	switch eventType {
	case DependencyEventTypes_Created:
		return DependencyCreated{Name: DependencyEventTypes_Created}, nil
	case DependencyEventTypes_Deleted:
		return DependencyDeleted{Name: DependencyEventTypes_Deleted}, nil
	case DependencyEventTypes_Updated:
		return DependencyUpdated{Name: DependencyEventTypes_Updated}, nil
	}
	return nil, fmt.Errorf("cannot parse event type '%s' for type 'Dependency'", eventType)
}
