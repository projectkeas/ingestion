package sdk

import (
	"encoding/json"
	"fmt"

	"github.com/projectkeas/ingestion/sdk/events"
)

type EventEnvelope struct {
	Metadata   EventMetadata   `json:"metadata"`
	Payload    interface{}     `json:"-"`
	rawPayload json.RawMessage `json:"payload"`
}

func (eventEnvelope *EventEnvelope) UnmarshalJSON(b []byte) error {
	var payload interface{}

	// Create fake type here so we don't get stuck in an infinite loop
	type envelope EventEnvelope

	// Do the initial parsing of the object to fill the metadata
	err := json.Unmarshal(b, (*envelope)(eventEnvelope))
	if err != nil {
		return err
	}

	// Now we should have enough to unmarshal the rest of the payload because of the metadata
	switch eventEnvelope.Metadata.EventType {
	case "Alert":
		payload, err = events.NewAlertEventFromType(eventEnvelope.Metadata.EventSubType, eventEnvelope.rawPayload)
	case "Artifact":
		payload, err = events.NewArtifactEventFromType(eventEnvelope.Metadata.EventSubType, eventEnvelope.rawPayload)
	case "Commit":
		payload, err = events.NewCommitEventFromType(eventEnvelope.Metadata.EventSubType, eventEnvelope.rawPayload)
	case "Dependency":
		payload, err = events.NewDependencyEventFromType(eventEnvelope.Metadata.EventSubType, eventEnvelope.rawPayload)
	case "Deployment":
		payload, err = events.NewDeploymentEventFromType(eventEnvelope.Metadata.EventSubType, eventEnvelope.rawPayload)
	case "Generic":
		payload = eventEnvelope.rawPayload
	case "Incident":
		payload, err = events.NewIncidentEventFromType(eventEnvelope.Metadata.EventSubType, eventEnvelope.rawPayload)
	case "PullRequest":
		payload, err = events.NewPullRequestEventFromType(eventEnvelope.Metadata.EventSubType, eventEnvelope.rawPayload)
	case "PullRequestComment":
		payload, err = events.NewPullRequestCommentEventFromType(eventEnvelope.Metadata.EventSubType, eventEnvelope.rawPayload)
	case "Repository":
		payload, err = events.NewRepositoryEventFromType(eventEnvelope.Metadata.EventSubType, eventEnvelope.rawPayload)
	case "SecurityAdvisory":
		payload, err = events.NewSecurityAdvisoryEventFromType(eventEnvelope.Metadata.EventSubType, eventEnvelope.rawPayload)
	case "Service":
		payload, err = events.NewServiceEventFromType(eventEnvelope.Metadata.EventSubType, eventEnvelope.rawPayload)
	case "WorkItem":
		payload, err = events.NewWorkItemEventFromType(eventEnvelope.Metadata.EventSubType, eventEnvelope.rawPayload)
	case "WorkItemComment":
		payload, err = events.NewWorkItemCommentEventFromType(eventEnvelope.Metadata.EventSubType, eventEnvelope.rawPayload)
	default:
		return fmt.Errorf("cannot parse event type '%s'", eventEnvelope.Metadata.EventType)
	}

	// If there's no error, ensure that we set the payload on the envelope
	if err != nil {
		return err
	}

	eventEnvelope.Payload = payload

	return nil
}
