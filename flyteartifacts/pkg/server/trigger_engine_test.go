package server

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestA(t *testing.T) {
	dur := ParseDuration("P1D")
	assert.Equal(t, time.Hour*24, dur)

}
