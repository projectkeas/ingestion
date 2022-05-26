package events

import (
	"encoding/json"
	"fmt"
)

const (
	PullRequestEventTypes_Created string = "Created"
	PullRequestEventTypes_Deleted string = "Deleted"
	PullRequestEventTypes_Updated string = "Updated"
)

type PullRequestCreated struct {
	Name string
}

type PullRequestDeleted struct {
	Name string
}

type PullRequestUpdated struct {
	Name string
}

func NewPullRequestEventFromType(eventType string, payload json.RawMessage) (interface{}, error) {
	switch eventType {
	case PullRequestEventTypes_Created:
		return PullRequestCreated{Name: PullRequestEventTypes_Created}, nil
	case PullRequestEventTypes_Deleted:
		return PullRequestDeleted{Name: PullRequestEventTypes_Deleted}, nil
	case PullRequestEventTypes_Updated:
		return PullRequestUpdated{Name: PullRequestEventTypes_Updated}, nil
	}
	return nil, fmt.Errorf("cannot parse event type '%s' for type 'PullRequest'", eventType)
}
