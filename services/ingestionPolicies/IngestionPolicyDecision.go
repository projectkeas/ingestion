package ingestionPolicies

type IngestionPolicyDecision struct {
	Allow bool
	TTL   int64
}
