package events

import (
	"encoding/json"
	"fmt"
)

const (
	WorkItemCommentEventTypes_Created string = "Created"
	WorkItemCommentEventTypes_Deleted string = "Deleted"
	WorkItemCommentEventTypes_Updated string = "Updated"
)

type WorkItemCommentCreated struct {
	Name string
}

type WorkItemCommentDeleted struct {
	Name string
}

type WorkItemCommentUpdated struct {
	Name string
}

func NewWorkItemCommentEventFromType(eventType string, payload json.RawMessage) (interface{}, error) {
	switch eventType {
	case WorkItemCommentEventTypes_Created:
		return WorkItemCommentCreated{Name: WorkItemCommentEventTypes_Created}, nil
	case WorkItemCommentEventTypes_Deleted:
		return WorkItemCommentDeleted{Name: WorkItemCommentEventTypes_Deleted}, nil
	case WorkItemCommentEventTypes_Updated:
		return WorkItemCommentUpdated{Name: WorkItemCommentEventTypes_Updated}, nil
	}
	return nil, fmt.Errorf("cannot parse event type '%s' for type 'WorkItemComment'", eventType)
}
