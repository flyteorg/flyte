package backoff

import (
	"sync/atomic"
	"time"
)

// AtomicTime represents an atomic.Value that stores time.Time.
type AtomicTime struct {
	v atomic.Value
}

// Loads the underlying time.Time.
func (a *AtomicTime) Load() time.Time {
	return a.v.Load().(time.Time)
}

// Stores time.Time to the underlying atomic.Value
func (a *AtomicTime) Store(t time.Time) {
	a.v.Store(t)
}

// Creates a new Atomic time.Time
func NewAtomicTime(t time.Time) AtomicTime {
	v := atomic.Value{}
	v.Store(t)

	return AtomicTime{
		v: v,
	}
}
