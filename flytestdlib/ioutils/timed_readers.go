package ioutils

import (
	"io"
	"io/ioutil"
)

// Defines a common interface for timers.
type Timer interface {
	// Stops the timer and reports observation.
	Stop() float64
}

func ReadAll(r io.Reader, t Timer) ([]byte, error) {
	defer t.Stop()
	return ioutil.ReadAll(r)
}
