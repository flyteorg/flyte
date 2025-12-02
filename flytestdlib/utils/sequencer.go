package utils

import (
	"sync"
	"sync/atomic"
)

// Sequencer is a thread-safe incremental integer counter. Note that it is a singleton, so
// GetSequencer.GetNext may not always start at 0.
type Sequencer interface {
	GetNext() uint64
	GetCur() uint64
}

type sequencer struct {
	val *uint64
}

var once sync.Once
var instance Sequencer

func GetSequencer() Sequencer {
	once.Do(func() {
		val := uint64(0)
		instance = &sequencer{val: &val}
	})
	return instance
}

// Get the next sequence number, 1 higher than the last and set it as the current one
func (s sequencer) GetNext() uint64 {
	x := atomic.AddUint64(s.val, 1)
	return x
}

// Get the current sequence number
func (s sequencer) GetCur() uint64 {
	return *s.val
}
