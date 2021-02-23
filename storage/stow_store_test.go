package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/graymeta/stow/google"
	"github.com/graymeta/stow/local"
	"github.com/pkg/errors"

	"github.com/graymeta/stow"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/internal/utils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
)

type mockStowLoc struct {
	stow.Location
	ContainerCb func(id string) (stow.Container, error)
}

func (m mockStowLoc) Container(id string) (stow.Container, error) {
	return m.ContainerCb(id)
}

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
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)

	const container = "container"
	t.Run("Happy Path", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		fn := fQNFn["s3"]
		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, false, testScope)
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
		fn := fQNFn["s3"]
		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, false, testScope)
		assert.NoError(t, err)
		err = s.WriteRaw(context.TODO(), DataReference("s3://container/path"), 3*MiB, Options{}, bytes.NewReader([]byte{}))
		assert.NoError(t, err)
		metadata, err := s.Head(context.TODO(), DataReference("s3://container/path"))
		assert.NoError(t, err)
		assert.True(t, metadata.Exists())
		_, err = s.ReadRaw(context.TODO(), DataReference("s3://container/path"))
		assert.Error(t, err)
		assert.True(t, IsExceedsLimit(err))
		assert.NotNil(t, errors.Cause(err))
	})

	t.Run("Happy Path multi-container enabled", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		fn := fQNFn["s3"]
		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				} else if id == "bad-container" {
					return newMockStowContainer("bad-container"), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, true, testScope)
		assert.NoError(t, err)
		err = s.WriteRaw(context.TODO(), "s3://bad-container/path", 0, Options{}, bytes.NewReader([]byte{}))
		assert.NoError(t, err)
		metadata, err := s.Head(context.TODO(), "s3://bad-container/path")
		if assert.NoError(t, err) {
			assert.True(t, metadata.Exists())
		}
		raw, err := s.ReadRaw(context.TODO(), "s3://bad-container/path")
		assert.NoError(t, err)
		rawBytes, err := ioutil.ReadAll(raw)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(rawBytes))
		assert.Equal(t, DataReference("s3://container"), s.GetBaseContainerFQN(context.TODO()))
	})

	t.Run("Happy Path multi-container bad", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		fn := fQNFn["s3"]
		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, true, testScope)
		assert.NoError(t, err)
		err = s.WriteRaw(context.TODO(), "s3://bad-container/path", 0, Options{}, bytes.NewReader([]byte{}))
		assert.Error(t, err)
		_, err = s.Head(context.TODO(), "s3://bad-container/path")
		assert.Error(t, err)
		_, err = s.ReadRaw(context.TODO(), "s3://bad-container/path")
		assert.Error(t, err)
	})
}

func TestNewLocalStore(t *testing.T) {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
	t.Run("Valid config", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		store, err := newStowRawStore(&Config{
			Stow: &StowConfig{
				Kind: local.Kind,
				Config: map[string]string{
					local.ConfigKeyPath: "./",
				},
			},
			InitContainer: "testdata",
		}, testScope.NewSubScope("x"))

		assert.NoError(t, err)
		assert.NotNil(t, store)

		// Stow local store expects the full path after the container portion (looks like a bug to me)
		rc, err := store.ReadRaw(context.TODO(), DataReference("file://testdata/config.yaml"))
		assert.NoError(t, err)
		if assert.NotNil(t, rc) {
			assert.NoError(t, rc.Close())
		}
	})

	t.Run("Invalid config", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		_, err := newStowRawStore(&Config{}, testScope)
		assert.Error(t, err)
	})

	t.Run("Initialize container", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		tmpDir, err := ioutil.TempDir("", "stdlib_local")
		assert.NoError(t, err)

		stats, err := os.Stat(tmpDir)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		store, err := newStowRawStore(&Config{
			Stow: &StowConfig{
				Kind: local.Kind,
				Config: map[string]string{
					local.ConfigKeyPath: tmpDir,
				},
			},
			InitContainer: "tmp",
		}, testScope.NewSubScope("y"))

		assert.NoError(t, err)
		assert.NotNil(t, store)

		stats, err = os.Stat(filepath.Join(tmpDir, "tmp"))
		assert.NoError(t, err)
		if assert.NotNil(t, stats) {
			assert.True(t, stats.IsDir())
		}
	})

	t.Run("missing init container", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		tmpDir, err := ioutil.TempDir("", "stdlib_local")
		assert.NoError(t, err)

		stats, err := os.Stat(tmpDir)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		store, err := newStowRawStore(&Config{
			Stow: &StowConfig{
				Kind: local.Kind,
				Config: map[string]string{
					local.ConfigKeyPath: tmpDir,
				},
			},
		}, testScope.NewSubScope("y"))

		assert.Error(t, err)
		assert.Nil(t, store)
	})

	t.Run("multi-container enabled", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		tmpDir, err := ioutil.TempDir("", "stdlib_local")
		assert.NoError(t, err)

		stats, err := os.Stat(tmpDir)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		store, err := newStowRawStore(&Config{
			Stow: &StowConfig{
				Kind: local.Kind,
				Config: map[string]string{
					local.ConfigKeyPath: tmpDir,
				},
			},
			InitContainer:         "tmp",
			MultiContainerEnabled: true,
		}, testScope.NewSubScope("y"))

		assert.NoError(t, err)
		assert.NotNil(t, store)

		stats, err = os.Stat(filepath.Join(tmpDir, "tmp"))
		assert.NoError(t, err)
		if assert.NotNil(t, stats) {
			assert.True(t, stats.IsDir())
		}
	})
}

func Test_newStowRawStore(t *testing.T) {
	type args struct {
		cfg          *Config
		metricsScope promutils.Scope
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"fail", args{&Config{}, promutils.NewTestScope()}, true},
		{"google", args{&Config{
			InitContainer: "flyte",
			Stow: &StowConfig{
				Kind: google.Kind,
				Config: map[string]string{
					google.ConfigProjectId: "x",
					google.ConfigScopes:    "y",
				},
			},
		}, promutils.NewTestScope()}, true},
		{"minio", args{&Config{
			Type:          TypeMinio,
			InitContainer: "some-container",
			Connection: ConnectionConfig{
				Endpoint: config.URL{URL: utils.MustParseURL("http://minio:9000")},
			},
		}, promutils.NewTestScope()}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newStowRawStore(tt.args.cfg, tt.args.metricsScope)
			if tt.wantErr {
				assert.Error(t, err, "newStowRawStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotNil(t, got, "Expected rawstore, found nil!")
		})
	}
}
