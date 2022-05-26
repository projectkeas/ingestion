package events

import (
	"encoding/json"
	"fmt"
)

const (
	RepositoryEventTypes_Created string = "Created"
	RepositoryEventTypes_Deleted string = "Deleted"
	RepositoryEventTypes_Updated string = "Updated"
)

type RepositoryCreated struct {
	Name string
}

type RepositoryDeleted struct {
	Name string
}

type RepositoryUpdated struct {
	Name string
}

func NewRepositoryEventFromType(eventType string, payload json.RawMessage) (interface{}, error) {
	switch eventType {
	case RepositoryEventTypes_Created:
		return RepositoryCreated{Name: RepositoryEventTypes_Created}, nil
	case RepositoryEventTypes_Deleted:
		return RepositoryDeleted{Name: RepositoryEventTypes_Deleted}, nil
	case RepositoryEventTypes_Updated:
		return RepositoryUpdated{Name: RepositoryEventTypes_Updated}, nil
	}
	return nil, fmt.Errorf("cannot parse event type '%s' for type 'Repository'", eventType)
}
