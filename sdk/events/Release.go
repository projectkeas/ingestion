package events

import (
	"encoding/json"
	"fmt"
)

const (
	ReleaseEventTypes_Created string = "Created"
	ReleaseEventTypes_Deleted string = "Deleted"
	ReleaseEventTypes_Updated string = "Updated"
)

type ReleaseCreated struct {
	Name string
}

type ReleaseDeleted struct {
	Name string
}

type ReleaseUpdated struct {
	Name string
}

func NewReleaseEventFromType(eventType string, payload json.RawMessage) (interface{}, error) {
	switch eventType {
	case ReleaseEventTypes_Created:
		return ReleaseCreated{Name: ReleaseEventTypes_Created}, nil
	case ReleaseEventTypes_Deleted:
		return ReleaseDeleted{Name: ReleaseEventTypes_Deleted}, nil
	case ReleaseEventTypes_Updated:
		return ReleaseUpdated{Name: ReleaseEventTypes_Updated}, nil
	}
	return nil, fmt.Errorf("cannot parse event type '%s' for type 'Release'", eventType)
}
