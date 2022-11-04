package catalog

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
)

//go:generate mockery -all -case=underscore

// Metadata to be associated with the catalog object
type Metadata struct {
	WorkflowExecutionIdentifier *core.WorkflowExecutionIdentifier
	NodeExecutionIdentifier     *core.NodeExecutionIdentifier
	TaskExecutionIdentifier     *core.TaskExecutionIdentifier
}

// An identifier for a catalog object.
type Key struct {
	Identifier     core.Identifier
	CacheVersion   string
	TypedInterface core.TypedInterface
	InputReader    io.InputReader
}

func (k Key) String() string {
	return fmt.Sprintf("%v:%v", k.Identifier, k.CacheVersion)
}

// Indicates that status of the query to Catalog. This can be returned for both Get and Put calls
type Status struct {
	cacheStatus core.CatalogCacheStatus
	metadata    *core.CatalogMetadata
}

func (s Status) GetCacheStatus() core.CatalogCacheStatus {
	return s.cacheStatus
}

func (s Status) GetMetadata() *core.CatalogMetadata {
	return s.metadata
}

func NewStatus(cacheStatus core.CatalogCacheStatus, md *core.CatalogMetadata) Status {
	return Status{cacheStatus: cacheStatus, metadata: md}
}

// Indicates the Entry in Catalog that was populated
type Entry struct {
	outputs io.OutputReader
	status  Status
}

func (e Entry) GetOutputs() io.OutputReader {
	return e.outputs
}

func (e Entry) GetStatus() Status {
	return e.status
}

func NewFailedCatalogEntry(status Status) Entry {
	return Entry{status: status}
}

func NewCatalogEntry(outputs io.OutputReader, status Status) Entry {
	return Entry{outputs: outputs, status: status}
}

// ReservationEntry encapsulates the current state of an artifact reservation within the catalog
type ReservationEntry struct {
	expiresAt         time.Time
	heartbeatInterval time.Duration
	ownerID           string
	status            core.CatalogReservation_Status
}

// Returns the expiration timestamp at which the reservation will no longer be valid
func (r ReservationEntry) GetExpiresAt() time.Time {
	return r.expiresAt
}

// Returns the heartbeat interval, denoting how often the catalog expects a reservation extension request
func (r ReservationEntry) GetHeartbeatInterval() time.Duration {
	return r.heartbeatInterval
}

// Returns the ID of the current reservation owner
func (r ReservationEntry) GetOwnerID() string {
	return r.ownerID
}

// Returns the status of the attempted reservation operation
func (r ReservationEntry) GetStatus() core.CatalogReservation_Status {
	return r.status
}

// Creates a new ReservationEntry using the status, all other fields are set to default values
func NewReservationEntryStatus(status core.CatalogReservation_Status) ReservationEntry {
	duration := 0 * time.Second
	return ReservationEntry{
		expiresAt:         time.Time{},
		heartbeatInterval: duration,
		ownerID:           "",
		status:            status,
	}
}

// Creates a new ReservationEntry populated with the specified parameters
func NewReservationEntry(expiresAt time.Time, heartbeatInterval time.Duration, ownerID string, status core.CatalogReservation_Status) ReservationEntry {
	return ReservationEntry{
		expiresAt:         expiresAt,
		heartbeatInterval: heartbeatInterval,
		ownerID:           ownerID,
		status:            status,
	}
}

// Client represents the default Catalog client that allows memoization and indexing of intermediate data in Flyte
type Client interface {
	// Get returns the artifact associated with the given key.
	Get(ctx context.Context, key Key) (Entry, error)
	// GetOrExtendReservation tries to retrieve a (valid) reservation for the given key, creating a new one using the
	// specified owner ID if none was found or updating an existing one if it has expired.
	GetOrExtendReservation(ctx context.Context, key Key, ownerID string, heartbeatInterval time.Duration) (*datacatalog.Reservation, error)
	// Put stores the given data using the specified key, creating artifact entries as required.
	// To update an existing artifact, use Update instead.
	Put(ctx context.Context, key Key, reader io.OutputReader, metadata Metadata) (Status, error)
	// Update updates existing data stored at the specified key, overwriting artifact entries with the new data provided.
	// To create a new (non-existent) artifact, use Put instead.
	Update(ctx context.Context, key Key, reader io.OutputReader, metadata Metadata) (Status, error)
	// ReleaseReservation releases an acquired reservation for the given key and owner ID.
	ReleaseReservation(ctx context.Context, key Key, ownerID string) error
}

func IsNotFound(err error) bool {
	taskStatus, ok := grpcStatus.FromError(err)
	return ok && taskStatus.Code() == codes.NotFound
}
