package events

import (
	"encoding/json"
	"fmt"
)

const (
	ArtifactEventTypes_Created string = "Created"
	ArtifactEventTypes_Deleted string = "Deleted"
	ArtifactEventTypes_Updated string = "Updated"
)

type ArtifactCreated struct {
	Name string
}

type ArtifactDeleted struct {
	Name string
}

type ArtifactUpdated struct {
	Name string
}

func NewArtifactEventFromType(eventType string, payload json.RawMessage) (interface{}, error) {
	switch eventType {
	case ArtifactEventTypes_Created:
		return ArtifactCreated{Name: ArtifactEventTypes_Created}, nil
	case ArtifactEventTypes_Deleted:
		return ArtifactDeleted{Name: ArtifactEventTypes_Deleted}, nil
	case ArtifactEventTypes_Updated:
		return ArtifactUpdated{Name: ArtifactEventTypes_Updated}, nil
	}
	return nil, fmt.Errorf("cannot parse event type '%s' for type 'Artifact'", eventType)
}
