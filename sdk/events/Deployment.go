package events

import (
	"encoding/json"
	"fmt"
)

const (
	DeploymentEventTypes_Created string = "Created"
	DeploymentEventTypes_Deleted string = "Deleted"
	DeploymentEventTypes_Updated string = "Updated"
)

type DeploymentCreated struct {
	Name string
}

type DeploymentDeleted struct {
	Name string
}

type DeploymentUpdated struct {
	Name string
}

func NewDeploymentEventFromType(eventType string, payload json.RawMessage) (interface{}, error) {
	switch eventType {
	case DeploymentEventTypes_Created:
		return DeploymentCreated{Name: DeploymentEventTypes_Created}, nil
	case DeploymentEventTypes_Deleted:
		return DeploymentDeleted{Name: DeploymentEventTypes_Deleted}, nil
	case DeploymentEventTypes_Updated:
		return DeploymentUpdated{Name: DeploymentEventTypes_Updated}, nil
	}
	return nil, fmt.Errorf("cannot parse event type '%s' for type 'Deployment'", eventType)
}
