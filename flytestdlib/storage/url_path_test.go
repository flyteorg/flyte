package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlPathConstructor_ConstructReference(t *testing.T) {
	s := URLPathConstructor{}
	r, err := s.ConstructReference(context.TODO(), DataReference("hello"), "key1", "key2/", "key3")
	assert.NoError(t, err)
	assert.Equal(t, "/hello/key1/key2/key3", r.String())
}
