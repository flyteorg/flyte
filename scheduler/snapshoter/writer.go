package snapshoter

import "io"

// Writer provides an interface to write the serialized form of the snapshot to a writer
type Writer interface {
	// WriteSnapshot writes the serialized form of the snapshot to the writer
	WriteSnapshot(writer io.Writer, snapshot Snapshot) error
}
