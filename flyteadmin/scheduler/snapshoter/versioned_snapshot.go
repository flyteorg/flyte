package snapshoter

import (
	"encoding/gob"
	"fmt"
	"io"
	"time"
)

// VersionedSnapshot stores the version and gob serialized form of the snapshot
// Provides a read and write methods to serialize and deserialize the gob format of the snapshot.
// Including a version provides compatibility check
type VersionedSnapshot struct {
	Version int
	Ser     []byte
}

func (s *VersionedSnapshot) WriteSnapshot(w io.Writer, snapshot Snapshot) error {
	byteContents, err := snapshot.Serialize()
	if err != nil {
		return err
	}
	s.Version = snapshot.GetVersion()
	s.Ser = byteContents
	enc := gob.NewEncoder(w)
	return enc.Encode(s)
}

func (s *VersionedSnapshot) ReadSnapshot(r io.Reader) (Snapshot, error) {
	err := gob.NewDecoder(r).Decode(s)
	if err != nil {
		return nil, err
	}
	if s.Version == 1 {
		snapShotV1 := SnapshotV1{LastTimes: map[string]*time.Time{}}
		err = snapShotV1.Deserialize(s.Ser)
		if err != nil {
			return nil, err
		}
		return &snapShotV1, nil
	}
	return nil, fmt.Errorf("unsupported version %v", s.Version)
}
