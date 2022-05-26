package events

import (
	"encoding/json"
	"fmt"
)

const (
	PullRequestCommentEventTypes_Created string = "Created"
	PullRequestCommentEventTypes_Deleted string = "Deleted"
	PullRequestCommentEventTypes_Updated string = "Updated"
)

type PullRequestCommentCreated struct {
	Name string
}

type PullRequestCommentDeleted struct {
	Name string
}

type PullRequestCommentUpdated struct {
	Name string
}

func NewPullRequestCommentEventFromType(eventType string, payload json.RawMessage) (interface{}, error) {
	switch eventType {
	case PullRequestCommentEventTypes_Created:
		return PullRequestCommentCreated{Name: PullRequestCommentEventTypes_Created}, nil
	case PullRequestCommentEventTypes_Deleted:
		return PullRequestCommentDeleted{Name: PullRequestCommentEventTypes_Deleted}, nil
	case PullRequestCommentEventTypes_Updated:
		return PullRequestCommentUpdated{Name: PullRequestCommentEventTypes_Updated}, nil
	}
	return nil, fmt.Errorf("cannot parse event type '%s' for type 'PullRequestComment'", eventType)
}
