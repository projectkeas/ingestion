package events

import (
	"encoding/json"
	"fmt"
)

const (
	CommitEventTypes_Created string = "Created"
)

type CommitCreated struct {
	Name string
}

func NewCommitEventFromType(eventType string, payload json.RawMessage) (interface{}, error) {
	switch eventType {
	case CommitEventTypes_Created:
		return CommitCreated{Name: CommitEventTypes_Created}, nil
	}
	return nil, fmt.Errorf("cannot parse event type '%s' for type 'Commit'", eventType)
}
