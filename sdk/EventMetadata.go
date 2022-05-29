package sdk

type EventMetadata struct {
	Source       string `json:"source" validate:"required,alpha,min=3,max=12"`
	EventType    string `json:"eventType" validate:"required,alpha"`
	EventSubType string `json:"eventSubType" validate:"required,alpha"`
	EventVersion string `json:"version" validate:"required,alpha"`
	EventTime    string `json:"eventTime" validate:"omitempty,datetime"`
	EventUUID    string `json:"eventUUID" validate:"omitempty,alphanum"`
}
