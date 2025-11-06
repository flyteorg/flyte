package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlPathConstructor_ConstructReference(t *testing.T) {
	s := URLPathConstructor{}
	t.Run("happy path", func(t *testing.T) {
		r, err := s.ConstructReference(context.TODO(), DataReference("hello"), "key1", "key2/", "key3")
		assert.NoError(t, err)
		assert.Equal(t, "/hello/key1/key2/key3", r.String())
	})

	t.Run("failed to parse base path", func(t *testing.T) {
		_, err := s.ConstructReference(context.TODO(), DataReference("*&^#&$@:%//"), "key1", "key2/", "key3")
		assert.Error(t, err)
	})

	t.Run("failed to parse nestedKeys", func(t *testing.T) {
		_, err := s.ConstructReference(context.TODO(), DataReference("*&^#&$@://"), "key1%", "key2/", "key3")
		assert.Error(t, err)
	})
}
