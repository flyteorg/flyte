package snapshoter

import (
	"bytes"
	"encoding/gob"
	"time"
)

// SnapshotV1 stores in the inmemory states of the schedules and there last execution timestamps.
// This map is created periodically from the jobstore of the gocron_wrapper and written to the DB.
// During bootup the serialized version of it is read from the DB and the schedules are initialized from it.
// V1 version so that in future if we add more fields for extending the functionality then this provides
// a backward compatible way to read old snapshots.
type SnapshotV1 struct {
	// LastTimes map of the schedule name to last execution timestamp
	LastTimes map[string]*time.Time
}

func (s *SnapshotV1) GetLastExecutionTime(key string) *time.Time {
	return s.LastTimes[key]
}

func (s *SnapshotV1) UpdateLastExecutionTime(key string, lastExecTime *time.Time) {
	// Load the last exec time for the schedule key and compare if its less than new LastExecTime
	// and only if it is then update the map
	s.LastTimes[key] = lastExecTime
}

func (s *SnapshotV1) Serialize() ([]byte, error) {
	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(s)
	return b.Bytes(), err
}

func (s *SnapshotV1) Deserialize(snapshot []byte) error {
	return gob.NewDecoder(bytes.NewBuffer(snapshot)).Decode(s)
}

func (s *SnapshotV1) IsEmpty() bool {
	return len(s.LastTimes) == 0
}

func (s *SnapshotV1) GetVersion() int {
	return 1
}

func (s *SnapshotV1) Create() Snapshot {
	return &SnapshotV1{
		LastTimes: map[string]*time.Time{},
	}
}
