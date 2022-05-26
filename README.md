# Ingestion

...

## Ingest Endpoint

Payload format:

```json
{
    "metadata": {
        "source": "Gitlab",                     // (Required) the source of the event
        "eventType": "PullRequest",             // (Required) the type of the event received
        "eventTime": "2022-01-01T13:33.00Z",    // (optional) the time the event occurred at the source. ISO-8601 formatted string, defaults to utc now
        "eventSubType": "Merged",               // (optional) provides further classification of the eventType field
        "eventUUID": "1234567567643",           // (optional) unique to the source
    },
    "payload": {
        // (required) variable structure based on source
    }
}
```

## EventTypes

- Alert
  - Triggered
  - Silenced
  - Resolved
- Artifact
  - Created
  - Deleted
  - Updated
- Commit
  - Created
- Dependency
  - Created
  - Deleted
  - Updated
- Deployment
  - Created
  - Deleted
  - Updated
- Incident
  - Created
  - Deleted
  - Updated
- PullRequest
  - Closed
  - Created
  - Merged
  - Updated
- PullRequestComment
  - Created
  - Deleted
  - Updated
- Release
  - Created
  - Deleted
  - Updated
- Repository
  - Created
  - Deleted
  - Updated
- SecurityAdvisory
  - Created
  - Deleted
  - Updated
- Service
  - Created
  - Deleted
  - Updated
- WorkItem
  - Created
  - Deleted
  - Updated
- WorkItemComment
  - Created
  - Deleted
  - Updated
