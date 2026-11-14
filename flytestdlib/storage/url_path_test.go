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

	t.Run("nestedKeys are literal path segments", func(t *testing.T) {
		for _, name := range []string{"a#b.csv", "what?.txt", "100%.csv", "50%25.csv", "t:1.csv"} {
			t.Run(name, func(t *testing.T) {
				r, err := s.ConstructReference(context.TODO(), DataReference("s3://bucket/prefix"), name)
				assert.NoError(t, err)
				_, container, key, err := r.Split()
				assert.NoError(t, err)
				assert.Equal(t, "bucket", container)
				assert.Equal(t, "prefix/"+name, key)
			})
		}
	})
}
