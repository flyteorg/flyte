package ioutils

import (
	"io"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBytesReadCloser(t *testing.T) {
	i := []byte("abc")
	r := NewBytesReadCloser(i)
	o, e := io.ReadAll(r)
	assert.NoError(t, e)
	assert.Equal(t, o, i)
	assert.NoError(t, r.Close())
}
