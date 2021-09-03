package snapshoter

import (
	"context"
)

// Persistence allows to read and save the serialized form of the snapshot from a storage.
// Currently we have DB implementation for it.
type Persistence interface {
	// Save Run(ctx context.Context)
	// Save saves the snapshot to the storage in a serialized form.
	Save(ctx context.Context, writer Writer, snapshot Snapshot)
	// Read reads the serialized snapshot from the storage and deserializes to its in memory format.
	Read(ctx context.Context, reader Reader) (Snapshot, error)
}
