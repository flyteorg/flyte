# [RFC] Cache Reservation API

**Authors:**

- @hamersaw 

## 1 Executive Summary

The cache reservation API provides a lease-based reservation system for cachable tasks in Flyte. The purpose is to mitigate simultaneous evaluations of cachable tasks over identical inputs, resulting in duplication of work and therefore inefficient resource utilization. The proposed approach will more effectively process workflows with potentially significant improvements in end-to-end workflow evaluation times for instances with long running cachable tasks.

## 2 Motivation

Currently, Flyte initializes cachable tasks with a lookup the the datacatalog cache. If a previous instance of the task (ie. identical version and inputs) has completed the cached values are used, otherwise the task is executed.

The issue is that disparate workflows, or unique executions of the same workflow, may execute an identical cachable task before a previous has completed. This results in multiple instances of the same task execution being performed simultaneously. For example, two workflows, namely A and B, contain the same long running (ex. 2 time units) identical cachable task. Workflow A executes the task beginning at t0 (finishing at t2) and workflow B executes the task at t1 (finishing at t3). The inefficiencies are twofold:

1. From t1 (workflow B task execution) to t2 (workflow A task completion) there are two instances of the same task performing identical work (albeit at different stages).
2. The execution from workflow B will not complete until t3, whereas it could use the cached results from workflow A at t2 to complete faster.

The proposed solution will mitigate unnecessary resource utilization by disallowing duplicate task executions and provide more efficient workflow processing by using all available cachable task results.


## Proposed Implementation

### Guide-level explanation

User-side functionality will require the introduction of a cacheReservation Flyte task annotation. An elementary example is how this may look is provided below:
    
```python
@task(cacheReservation=true)
def long_running_task():
    # redacted long operation
    
@workflow
def wf():
    return long_running_task()
```

The cacheReservation tag is similar to the existing cache tag, but will take precendence over it. Meaning if both are provided on a specific task, caching will be performed using the cache reservation API.

Users employing this task will be oblivious to the backend implementation. However, they should be aware that it will result in minimal additional overhead. Cache reservations require additional message passing and periodic task and reservation monitoring. Therefore, relatively short running tasks may be better fit for the existing cache mechanism.

### Flyte IDL

We need to add gRPC calls into `flyteidl` allowing the acquisition and release of artifact reservations. This functionality has already been merged into `flyteidl`.  The proposed gRPC and protobuf message definitions are provided below:

```protobuf
service DataCatalog {
    // ... redacted gRPC calls
    
    // Get an artifact and the corresponding data. If the artifact does not exist,
    // try to reserve a spot for populating the artifact.
    // Once you preserve a spot, you should peridically extend the reservation before expiration
    // with an identical call. Otherwise, the reservation may be acquired by another task.
    // If the same owner_id calls this API for the same dataset and it has an active reservation and the artifacts have not been written yet by a different owner, the API will respond with an Acquired Reservation Status (providing idempotency).
    // Note: We may have multiple concurrent tasks with the same signature
    // and the same input that try to populate the same artifact at the same time.
    // Thus with reservation, only one task can run at a time, until the reservation
    // expires.
    // Note: If task A does not extend the reservation in time and the reservation
    // expires, another task B may take over the reservation, resulting in two tasks
    // A and B running in parallel. So a third task C may get the Artifact from A or B,
    // whichever writes last.
    rpc GetOrReserveArtifact (GetOrReserveArtifactRequest) returns (GetOrReserveArtifactResponse);

    // Release the reservation when the task holding the spot fails so that the other tasks
    // can grab the spot.
    rpc ReleaseReservation (ReleaseReservationRequest) returns (ReleaseReservationResponse);
}
    
/*
 * ReservationID message that is composed of several string fields.
 */
message ReservationID {
    DatasetID dataset_id = 1;
    string tag_name = 2;
}

// Get the Artifact or try to reserve a spot if the Artifact does not exist.
message GetOrReserveArtifactRequest {
    ReservationID reservation_id = 1;
    string owner_id = 2;
}

// The status of a reservation including owner, state, expiration timestamp, and various metadata.
message ReservationStatus {
    ReservationID reservation_id = 1;
    string owner_id = 2;
    State state = 3;
    google.protobuf.Timestamp expires_at = 4; // Expiration timestamp of this reservation
    google.protobuf.Duration heartbeat_interval = 5; // Recommended heartbeat interval to extend reservation
    Metadata metadata = 6;

    enum State {
        // Acquired the reservation successfully.
        ACQUIRED = 0;

        // Indicates an existing active reservation exist for a different owner_id.
        ALREADY_IN_PROGRESS = 1;
    };
}

// Response to get artifact or reserve spot.
message GetOrReserveArtifactResponse {
    oneof value {
        Artifact artifact = 1;
        ReservationStatus reservation_status = 2;
    }
}

// Request to release reservation
message ReleaseReservationRequest {
    ReservationID reservation_id = 1;
    string owner_id = 2;
}

// Response to release reservation
message ReleaseReservationResponse {
}
```

In this [Pull Request](https://github.com/flyteorg/flyteidl/pull/215) we present these changes by adding the ReservationID message and removing the existing ExtendReservation gRPC call since reservation extensions are now handled as part of the GetOrReserveArtifact call

### Data Catalog

The `datacatalog` service will be responsible for managing cache reservations. This will entail the addition of a new ReservationManager and ReservationRepo (with gorm implementation) per the project standards. Additionally it requires a new table in the db where reservations are uniquely defined based on DatasetID and an artifact tag. The model definition is provided below:

```go=
// ReservationKey uniquely identifies a reservation
type ReservationKey struct {
	DatasetProject string `gorm:"primary_key"`
	DatasetName    string `gorm:"primary_key"`
	DatasetDomain  string `gorm:"primary_key"`
	DatasetVersion string `gorm:"primary_key"`
	TagName        string `gorm:"primary_key"`
}

// Reservation tracks the metadata needed to allow
// task cache serialization
type Reservation struct {
	BaseModel
	ReservationKey

	// Identifies who owns the reservation
	OwnerID string

	// When the reservation will expire
	ExpiresAt          time.Time
	SerializedMetadata []byte
}
```

All database operations are performed with write consistency, where records are only inserted or updated on restrictive conditions. This eliminates the possibility for race conditions. Where two executions attempt to aquire a cache reservation simultaneously, only one can succeeed.

Additionally, the `datacatalog` configuration file defines heartbeat-interval-seconds and heartbeat-grace-period-multiplier to define the expected heartbeat interval of reservation extensions and set the reservation expiration (computed as heartbeat-interval-seconds * heartbeat-grace-period-multiplier).

All of the aforemented functionality has been implemented in this [Pull Request](https://github.com/flyteorg/datacatalog/pull/47).

### Flyte Propeller

`flytepropeller` implementation details are under active consideration. The control flow logic for a task with cacheReservation enabled is as follows:

1. Attempt to acquire an artifact reservation with GetOrReserveArtifact which returns:
    - Artifact: artifact has already been cached, use the provided results (complete)
    - AlreadyInProgress: another task execution already holds the reservation, wait and check again later (GOTO 1)
    - Acquired: successfully acquired reservation, begin task execution (GOTO 2)
2. Periodically monitor task execution in control loop resulting in either:
    - Running: task still running, extend artifact reservation (GOTO 2)
    - Completed: task is completed, add result artifact to `datacatalog` cache and release artifact reservation (completed)
    
Task execution failures are handled with reservation expirations. If a task which holds the artifact reservation fails, another task will acquire the reservation in the periodic call to GetOrReserveArtifact.

## 4 Metrics & Dashboards

- Latency of reservation gRPC function calls to get an idea of the overhead
- Task execution idle time in comparison to heartbeat-interval-seconds configuration

## 5 Drawbacks

The advantages / disadvantages may not be clear to users. Intuitively, this feature may be viewed as a replacement of the existing cache mechanism. It needs to be explicitely stated this is not the case.

Performance comparisons with the existing cache mechanism may be difficult to interpret. The massive diversity in tasks make it hard to implicitely identify scenarios where this solution would improve performance.

## 6 Alternatives

A reservation management system is the only obvious solution to enable different task executions to be aware of identical computation.

## 7 Potential Impact and Dependencies

A malicious entity could eternally extend a cache reservation without performing any computation. This would effectively block all tasks relying on the completion of the cached task execution.

## 8 Unresolved questions

- Definition of the reservation heartbeat interval and expiration grace period duration are defined within the `datacatalog`. Are these configuration values something that can remain static between tasks? Or are they better provided as part of the reservation request?
- `flytepropeller` integration details are expected to resolve through the implementation of this feature. A brief list of unclear details are provided here:
    - Does the control loop paradigm ensure reservation extensions before an expiration?

## 9 Conclusion

This solution for reserving artifact caches will mitigate unecessary resource utilization for cachable, long running tasks. It is designed for scale to cope with large deployments and effectively manages reservation management including reservation expirations and race conditions during acquisition. It has the potential for significant performance improvements in disparate workflows, or sequential executions of the same workflow, where an expensive, cachable task is continuously executed.