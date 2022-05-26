package events

import (
	"encoding/json"
	"fmt"
)

const (
	AlertEventTypes_Triggered string = "Triggered"
	AlertEventTypes_Silenced  string = "Silenced"
	AlertEventTypes_Resolved  string = "Resolved"
)

type AlertTriggered struct {
	Name string
}

type AlertSilenced struct {
	Name string
}

type AlertResolved struct {
	Name string
}

func NewAlertEventFromType(eventType string, payload json.RawMessage) (interface{}, error) {
	switch eventType {
	case AlertEventTypes_Triggered:
		return AlertTriggered{Name: AlertEventTypes_Triggered}, nil
	case AlertEventTypes_Silenced:
		return AlertSilenced{Name: AlertEventTypes_Silenced}, nil
	case AlertEventTypes_Resolved:
		return AlertResolved{Name: AlertEventTypes_Resolved}, nil
	}
	return nil, fmt.Errorf("cannot parse event type '%s' for type 'Alert'", eventType)
}
