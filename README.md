# Ingestion

The Keas ingestion API validates incoming event data against known schemas before applying one or more ingestion policies against the request. Consumers are free to define one of two [CRDs](github.com/projectkeas/crds):

- [EventType](https://github.com/projectkeas/crds/blob/main/manifests/keas.io_eventtypes.yaml): A versioned event schema. An EventType must be added for the system to accept the request.
- [IngestionPolicy](https://github.com/projectkeas/crds/blob/main/manifests/keas.io_ingestionpolicies.yaml): A policy to determine whether certain events should be stored in the system and for how long.

The ingestion API watches both resource types for any changes and reflects them immediately in the API. Every 2 minutes the system will perform a cache sync in the case of a network partition and ensure the consistency of all CRDs registered.

## Endpoints

|Url|Methods|Description|Payload|
|---|---|---|---|
|`/ingest`|POST|Captures a given event into the system (assuming it passes validation and ingestion policies)|[link](#ingest-payload)|
|`/_system/health`|GET|The liveness health check endpoint||
|`/_system/health/ready`|GET|The readiness health check endpoint||

The `/_system/*` endpoints are anonymous but all other endpoints have authentication in the format `Authorization: ApiKey <value from secret ingestion-secret>`

### Ingest Payload

```json
{
    "metadata": {
        "version": "1.0.0",                     // (Required) The version of the schema. Format: <major>.<minor>.<version>
        "source": "Gitlab",                     // (Required) The source of the event. Format: ^[A-z\-]{3,63}$
        "type": "PullRequest",                  // (Required) The type of the event received. Format: ^[A-z\-]{3,63}$
        "subType": "Merged",                    // (Optional) Provides further classification of the eventType field. Format: 
        "eventTime": "2022-01-01T13:33.00Z",    // (Optional) The time the event occurred at the source. ISO-8601 formatted string, defaults to utc now
        "eventUUID": "1234567567643",           // (Optional) A unique identifier of the event, usually coming from the source system
    },
    "payload": {
        // (required) variable structure based on source
    }
}
```

Note: Any other metadata fields than those listed above will be rejected by the system

### Error Response

During the course of development, you may receive one or more of the reason codes listed below:

|Reason|Description|Fix|
|---|---|---|
|request-body|The system was unable to parse|Ensure that the body is a valid JSON object|
|metadata-missing|The system could not get the property `metadata` from the request|Ensure that the field is present and lowercased|
|payload-missing|The system could not get the property `payload` from the request|Ensure that the field is present and lowercased|
|metadata|The supplied metadata did not match the schema `metadata`|Ensure that the `metadata` section is formatted according to [this schema](https://github.com/projectkeas/ingestion/blob/e1c07265b7799cebf47f5c296f4f149b6b5372fa/handlers/ingestionHandler/handler.go#L20-L54)|
|metadata-failure|There was a server side error deserializing the `metadata` section|N/A|
|event-validation|The system failed to successfully validate the request payload against the one stored in the system|Ensure that the event sent to the system matches the schema registered|
|event-validation-failure|There was a server side error |N/A|
|ingestion-service-failure|There was a server side error whilst processing one or more ingestion policies|Ensure all ingestion policies registered are valid [Rego policies](https://www.openpolicyagent.org/docs/latest/policy-language/)|
|ingestion-service-rejected|One or more policies evaluated the ingestion policy as disallowing the request|Adjust the ingestion policy if deemed that the policy is incorrect otherwise - N/A|

## Configuration

The ingestion system looks for two required configuration objects within a Kubernetes cluster:

- ConfigMap: `ingestion-cm`
- Secret: `ingestion-secret`

The readiness check will fail if the secret `ingestion-secret` is missing as the server requires a token for the NATS cluster and a ApiKey to use for authenticating users.

Example configurations:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ingestion-cm
data:
  log.level: debug
  nats.address: '10.0.0.31'
  nats.port: '4222'
---
apiVersion: v1
kind: Secret
metadata:
  name: ingestion-secret
stringData:
  auth.token: Testing!
  nats.token: Testing!
```
