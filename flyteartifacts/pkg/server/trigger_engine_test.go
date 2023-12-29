package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestA(t *testing.T) {
	dur := ParseDuration("P1D")
	assert.Equal(t, time.Hour*24, dur)

}
