package snapshoter

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadWriteSnapshot(t *testing.T) {
	t.Run("successful read write", func(t *testing.T) {
		var bytesArray []byte
		f := bytes.NewBuffer(bytesArray)
		writer := VersionedSnapshot{}
		snapshot := &SnapshotV1{
			LastTimes: map[string]*time.Time{},
		}
		currTime := time.Now()
		snapshot.LastTimes["schedule1"] = &currTime
		err := writer.WriteSnapshot(f, snapshot)
		assert.Nil(t, err)
		r := bytes.NewReader(f.Bytes())
		reader := VersionedSnapshot{}
		s, err := reader.ReadSnapshot(r)
		assert.Nil(t, err)
		assert.Equal(t, s.IsEmpty(), false)
		assert.NotNil(t, s.GetLastExecutionTime("schedule1"))
	})

	t.Run("successful write unsuccessful read", func(t *testing.T) {
		var bytesArray []byte
		f := bytes.NewBuffer(bytesArray)
		writer := VersionedSnapshot{}
		snapshot := &SnapshotV1{
			LastTimes: map[string]*time.Time{},
		}
		currTime := time.Now()
		snapshot.LastTimes["schedule1"] = &currTime
		err := writer.WriteSnapshot(f, snapshot)
		assert.Nil(t, err)

		bytesArray = f.Bytes()
		bytesArray[len(bytesArray)-1] = 1
		r := bytes.NewReader(f.Bytes())
		reader := VersionedSnapshot{}
		_, err = reader.ReadSnapshot(r)
		assert.NotNil(t, err)
	})
}
