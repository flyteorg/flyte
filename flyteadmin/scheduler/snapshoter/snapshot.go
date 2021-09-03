package snapshoter

import (
	"time"
)

// Snapshot used by the scheduler for creating, updating and reading snapshots of the schedules.
type Snapshot interface {
	// GetLastExecutionTime of the schedule given by the key
	GetLastExecutionTime(key string) *time.Time
	// UpdateLastExecutionTime of the schedule given by key to the lastExecTime
	UpdateLastExecutionTime(key string, lastExecTime *time.Time)
	// CreateSnapshot creates the snapshot of all the schedules and there execution times.
	Serialize() ([]byte, error)
	// BootstrapFrom bootstraps the snapshot from a byte array
	Deserialize(snapshot []byte) error
	// GetVersion gets the version number of snapshot written
	GetVersion() int
	// IsEmpty returns true if the snapshot contains no schedules
	IsEmpty() bool
	// Create an empty snapshot
	Create() Snapshot
}
