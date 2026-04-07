package ioutils

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

func TestReadAll(t *testing.T) {
	r := bytes.NewReader([]byte("hello"))
	s := promutils.NewTestScope()
	w, e := s.NewStopWatch("x", "empty", time.Millisecond)
	assert.NoError(t, e)
	b, err := ReadAll(r, w.Start())
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(b))
}
