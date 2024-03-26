package snapshoter

import "io"

// Reader provides an interface to read the snapshot and deserialize it to its in memory format.
type Reader interface {
	// ReadSnapshot reads the snapshot from the reader
	ReadSnapshot(reader io.Reader) (Snapshot, error)
}
