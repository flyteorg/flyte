package ioutils

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBytesReadCloser(t *testing.T) {
	i := []byte("abc")
	r := NewBytesReadCloser(i)
	o, e := ioutil.ReadAll(r)
	assert.NoError(t, e)
	assert.Equal(t, o, i)
	assert.NoError(t, r.Close())
}
