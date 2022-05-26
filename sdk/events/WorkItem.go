package events

import (
	"encoding/json"
	"fmt"
)

const (
	WorkItemEventTypes_Created string = "Created"
	WorkItemEventTypes_Deleted string = "Deleted"
	WorkItemEventTypes_Updated string = "Updated"
)

type WorkItemCreated struct {
	Name string
}

type WorkItemDeleted struct {
	Name string
}

type WorkItemUpdated struct {
	Name string
}

func NewWorkItemEventFromType(eventType string, payload json.RawMessage) (interface{}, error) {
	switch eventType {
	case WorkItemEventTypes_Created:
		return WorkItemCreated{Name: WorkItemEventTypes_Created}, nil
	case WorkItemEventTypes_Deleted:
		return WorkItemDeleted{Name: WorkItemEventTypes_Deleted}, nil
	case WorkItemEventTypes_Updated:
		return WorkItemUpdated{Name: WorkItemEventTypes_Updated}, nil
	}
	return nil, fmt.Errorf("cannot parse event type '%s' for type 'WorkItem'", eventType)
}
