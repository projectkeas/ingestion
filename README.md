# Ingestion

The Keas ingestion API validates incoming event data against known schemas before applying one or more ingestion policies against the request. Consumers are free to define one of two [CRDs](github.com/projectkeas/crds):

- [EventType](https://github.com/projectkeas/crds/blob/main/manifests/keas.io_eventtypes.yaml): A versioned event schema. An EventType must be added for the system to accept the request.
- [IngestionPolicy](https://github.com/projectkeas/crds/blob/main/manifests/keas.io_ingestionpolicies.yaml): A policy to determine whether certain events should be stored in the system and for how long.

The ingestion API watches both resource types for any changes and reflects them immediately in the API. Every 2 minutes the system will perform a cache sync in the case of a network partition and ensure the consistency of all CRDs registered. All events processed by this API must adhere to the [CloudEvents 1.0 standard](https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/spec.md).

## Endpoints

|Url|Methods|Description|Payload|
|---|---|---|---|
|`/ingest`|POST|Captures a given event into the system (assuming it passes validation and ingestion policies)|[link](#ingest-payload)|
|`/_system/health`|GET|The liveness health check endpoint||
|`/_system/health/ready`|GET|The readiness health check endpoint||

The `/_system/*` endpoints are anonymous but all other endpoints have authentication in the format `Authorization: ApiKey <value from secret ingestion-secret>`

### Error Response

During the course of development, you may receive one or more of the reason codes listed below:

|Reason|Description|Fix|
|---|---|---|
|request-body|The system was unable to parse|Ensure that the body is a valid JSON object|
|cloud-event-validation|The server failed to receive a valid cloud event|See the errors property of the response|
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
  ingestion.auth.token: Testing!
```
