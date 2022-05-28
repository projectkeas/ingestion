package sdk

type EventEnvelope struct {
	Metadata EventMetadata          `json:"metadata"`
	Payload  map[string]interface{} `json:"payload"`
}
