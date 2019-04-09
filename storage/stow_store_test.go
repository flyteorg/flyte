package storage

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/url"
	"testing"
	"time"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/graymeta/stow"
	"github.com/stretchr/testify/assert"
)

type mockStowContainer struct {
	id    string
	items map[string]mockStowItem
}

func (m mockStowContainer) ID() string {
	return m.id
}

func (m mockStowContainer) Name() string {
	return m.id
}

func (m mockStowContainer) Item(id string) (stow.Item, error) {
	if item, found := m.items[id]; found {
		return item, nil
	}

	return nil, stow.ErrNotFound
}

func (mockStowContainer) Items(prefix, cursor string, count int) ([]stow.Item, string, error) {
	return []stow.Item{}, "", nil
}

func (mockStowContainer) RemoveItem(id string) error {
	return nil
}

func (m *mockStowContainer) Put(name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error) {
	item := mockStowItem{url: name, size: size}
	m.items[name] = item
	return item, nil
}

func newMockStowContainer(id string) *mockStowContainer {
	return &mockStowContainer{
		id:    id,
		items: map[string]mockStowItem{},
	}
}

type mockStowItem struct {
	url  string
	size int64
}

func (m mockStowItem) ID() string {
	return m.url
}

func (m mockStowItem) Name() string {
	return m.url
}

func (m mockStowItem) URL() *url.URL {
	u, err := url.Parse(m.url)
	if err != nil {
		panic(err)
	}

	return u
}

func (m mockStowItem) Size() (int64, error) {
	return m.size, nil
}

func (mockStowItem) Open() (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewReader([]byte{})), nil
}

func (mockStowItem) ETag() (string, error) {
	return "", nil
}

func (mockStowItem) LastMod() (time.Time, error) {
	return time.Now(), nil
}

func (mockStowItem) Metadata() (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func TestStowStore_ReadRaw(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		s, err := NewStowRawStore(s3FQN("container"), newMockStowContainer("container"), testScope)
		assert.NoError(t, err)
		err = s.WriteRaw(context.TODO(), DataReference("s3://container/path"), 0, Options{}, bytes.NewReader([]byte{}))
		assert.NoError(t, err)
		metadata, err := s.Head(context.TODO(), DataReference("s3://container/path"))
		assert.NoError(t, err)
		assert.True(t, metadata.Exists())
		raw, err := s.ReadRaw(context.TODO(), DataReference("s3://container/path"))
		assert.NoError(t, err)
		rawBytes, err := ioutil.ReadAll(raw)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(rawBytes))
		assert.Equal(t, DataReference("s3://container"), s.GetBaseContainerFQN(context.TODO()))
	})

	t.Run("Exceeds limit", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		s, err := NewStowRawStore(s3FQN("container"), newMockStowContainer("container"), testScope)
		assert.NoError(t, err)
		err = s.WriteRaw(context.TODO(), DataReference("s3://container/path"), 3*MiB, Options{}, bytes.NewReader([]byte{}))
		assert.NoError(t, err)
		metadata, err := s.Head(context.TODO(), DataReference("s3://container/path"))
		assert.NoError(t, err)
		assert.True(t, metadata.Exists())
		_, err = s.ReadRaw(context.TODO(), DataReference("s3://container/path"))
		assert.Error(t, err)
		assert.True(t, IsExceedsLimit(err))
	})
}
