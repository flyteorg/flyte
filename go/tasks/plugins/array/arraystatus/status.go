/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package arraystatus

import (
	"encoding/binary"
	"hash/fnv"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flytestdlib/bitarray"
)

type JobID = string
type ArraySummary map[core.Phase]int64

type ArrayStatus struct {
	// Summary of the array job. It's a map of phases and how many jobs are in that phase.
	Summary ArraySummary `json:"summary"`

	// Status of every job in the array.
	Detailed bitarray.CompactArray `json:"details"`
}

// HashCode computes a hash of the phase indicies stored in the Detailed array to uniquely represent
// a collection of subtask phases.
func (a ArrayStatus) HashCode() (uint64, error) {
	hash := fnv.New64()
	bytes := make([]byte, 8)
	for _, phaseIndex := range a.Detailed.GetItems() {
		binary.LittleEndian.PutUint64(bytes, phaseIndex)
		_, err := hash.Write(bytes)
		if err != nil {
			return 0, err
		}
	}

	return hash.Sum64(), nil
}

// This is a status object that is returned after we make Catalog calls to see if subtasks are Cached
type ArrayCachedStatus struct {
	CachedJobs *bitarray.BitSet `json:"cachedJobs"`
	NumCached  uint             `json:"numCached"`
}

func deleteOrSet(summary ArraySummary, key core.Phase, value int64) {
	if value == 0 {
		delete(summary, key)
	} else {
		summary[key] = value
	}
}

func (in ArraySummary) IncByCount(phase core.Phase, count int64) {
	if existing, found := in[phase]; !found {
		in[phase] = count
	} else {
		in[phase] = existing + count
	}
}

func (in ArraySummary) Inc(phase core.Phase) {
	in.IncByCount(phase, 1)
}

func (in ArraySummary) Dec(phase core.Phase) {
	// TODO: Error if already 0?
	in.IncByCount(phase, -1)
}

func (in ArraySummary) MergeFrom(other ArraySummary) (updated bool) {
	// TODO: Refactor using sets
	if other == nil {
		for key := range in {
			delete(in, key)
			updated = true
		}

		return
	}

	for key, otherValue := range other {
		if value, found := in[key]; found {
			if value != otherValue {
				deleteOrSet(in, key, otherValue)
				updated = true
			}
		} else if otherValue != 0 {
			in[key] = otherValue
			updated = true
		}
	}

	for key := range in {
		if _, found := other[key]; !found {
			delete(in, key)
		}
	}

	return
}
