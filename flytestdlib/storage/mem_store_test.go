package storage

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryStore_Head(t *testing.T) {
	t.Run("Empty store", func(t *testing.T) {
		s, err := NewInMemoryRawStore(context.TODO(), &Config{}, metrics)
		assert.NoError(t, err)
		metadata, err := s.Head(context.TODO(), DataReference("hello"))
		assert.NoError(t, err)
		assert.False(t, metadata.Exists())
	})

	t.Run("Existing Item", func(t *testing.T) {
		s, err := NewInMemoryRawStore(context.TODO(), &Config{}, metrics)
		assert.NoError(t, err)
		err = s.WriteRaw(context.TODO(), DataReference("hello"), 0, Options{}, bytes.NewReader([]byte{}))
		assert.NoError(t, err)

		metadata, err := s.Head(context.TODO(), DataReference("hello"))
		assert.NoError(t, err)
		assert.True(t, metadata.Exists())
	})
}

func TestInMemoryStore_GetItems(t *testing.T) {
	t.Run("Nil Path", func(t *testing.T) {
		s, err := NewInMemoryRawStore(context.TODO(), &Config{}, metrics)
		assert.NoError(t, err)

		items, err := s.GetItems(context.TODO(), DataReference("hello"))
		assert.Error(t, err)
		assert.Nil(t, items)
	})

	t.Run("No Items", func(t *testing.T) {
		s, err := NewInMemoryRawStore(context.TODO(), &Config{}, metrics)
		assert.NoError(t, err)
		err = s.WriteRaw(context.TODO(), DataReference("folder"), 0, Options{}, bytes.NewReader([]byte{}))
		assert.NoError(t, err)

		items, err := s.GetItems(context.TODO(), DataReference("folder"))
		assert.Error(t, err)
		assert.Nil(t, items)
	})

	t.Run("Existing Items", func(t *testing.T) {
		s, err := NewInMemoryRawStore(context.TODO(), &Config{}, metrics)
		assert.NoError(t, err)
		err = s.WriteRaw(context.TODO(), DataReference("folder/file1"), 0, Options{}, bytes.NewReader([]byte{}))
		assert.NoError(t, err)

		err = s.WriteRaw(context.TODO(), DataReference("folder/file2"), 0, Options{}, bytes.NewReader([]byte{}))
		assert.NoError(t, err)

		items, err := s.GetItems(context.TODO(), DataReference("folder"))
		assert.NoError(t, err)
		assert.Equal(t, 2, len(items))
		assert.Equal(t, "folder/file1", items[0])
		assert.Equal(t, "folder/file2", items[1])
	})
}

func TestInMemoryStore_ReadRaw(t *testing.T) {
	t.Run("Empty store", func(t *testing.T) {
		s, err := NewInMemoryRawStore(context.TODO(), &Config{}, metrics)
		assert.NoError(t, err)

		raw, err := s.ReadRaw(context.TODO(), DataReference("hello"))
		assert.Error(t, err)
		assert.Nil(t, raw)
	})

	t.Run("Existing Item", func(t *testing.T) {
		s, err := NewInMemoryRawStore(context.TODO(), &Config{}, metrics)
		assert.NoError(t, err)

		err = s.WriteRaw(context.TODO(), DataReference("hello"), 0, Options{}, bytes.NewReader([]byte{}))
		assert.NoError(t, err)

		_, err = s.ReadRaw(context.TODO(), DataReference("hello"))
		assert.NoError(t, err)
	})
}

func TestInMemoryStore_Clear(t *testing.T) {
	m, err := NewInMemoryRawStore(context.TODO(), &Config{}, metrics)
	assert.NoError(t, err)

	mStore := m.(*InMemoryStore)
	err = m.WriteRaw(context.TODO(), DataReference("hello"), 0, Options{}, bytes.NewReader([]byte("world")))
	assert.NoError(t, err)

	_, err = m.ReadRaw(context.TODO(), DataReference("hello"))
	assert.NoError(t, err)

	err = mStore.Clear(context.TODO())
	assert.NoError(t, err)

	_, err = m.ReadRaw(context.TODO(), DataReference("hello"))
	assert.Error(t, err)
	assert.True(t, IsNotFound(err))
}

func TestInMemoryStore_Delete(t *testing.T) {
	m, err := NewInMemoryRawStore(context.TODO(), &Config{}, metrics)
	assert.NoError(t, err)

	mStore := m.(*InMemoryStore)
	err = m.WriteRaw(context.TODO(), DataReference("hello"), 0, Options{}, bytes.NewReader([]byte("world")))
	assert.NoError(t, err)

	_, err = m.ReadRaw(context.TODO(), DataReference("hello"))
	assert.NoError(t, err)

	err = mStore.Delete(context.TODO(), DataReference("hello"))
	assert.NoError(t, err)

	_, err = m.ReadRaw(context.TODO(), DataReference("hello"))
	assert.Error(t, err)
	assert.True(t, IsNotFound(err))

	err = mStore.Delete(context.TODO(), DataReference("hello"))
	assert.Error(t, err)
	assert.True(t, IsNotFound(err))
}
