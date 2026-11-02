package storage

import (
	"bytes"
	"context"
	errors2 "errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/aws/smithy-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils/labeled"
	"github.com/flyteorg/stow"
	"github.com/flyteorg/stow/azure"
	"github.com/flyteorg/stow/google"
	"github.com/flyteorg/stow/local"
	"github.com/flyteorg/stow/oracle"
	"github.com/flyteorg/stow/s3"
	"github.com/flyteorg/stow/swift"
)

type mockStowLoc struct {
	stow.Location
	ContainerCb       func(id string) (stow.Container, error)
	CreateContainerCb func(name string) (stow.Container, error)
}

func (m mockStowLoc) Container(id string) (stow.Container, error) {
	return m.ContainerCb(id)
}

func (m mockStowLoc) CreateContainer(name string) (stow.Container, error) {
	return m.CreateContainerCb(name)
}

type mockStowContainer struct {
	id    string
	items map[string]mockStowItem
	putCB func(name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error)
}

// CreateSignedURL creates a signed url with the provided properties.
func (m mockStowContainer) PreSignRequest(_ context.Context, _ stow.ClientMethod, s string,
	_ stow.PresignRequestParams) (response stow.PresignResponse, err error) {
	return stow.PresignResponse{Url: s}, nil
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

func (m mockStowContainer) Items(prefix, cursor string, count int) ([]stow.Item, string, error) {
	startIndex := 0
	if cursor != "" {
		index, err := strconv.Atoi(cursor)
		if err != nil {
			return nil, "", fmt.Errorf("Invalid cursor '%s'", cursor)
		}
		startIndex = index
	}
	endIndexExc := min(len(m.items), startIndex+count)

	itemKeys := make([]string, len(m.items))
	index := 0
	for key := range m.items {
		itemKeys[index] = key
		index++
	}
	sort.Strings(itemKeys)

	numItems := endIndexExc - startIndex
	results := make([]stow.Item, numItems)
	for index, itemKey := range itemKeys[startIndex:endIndexExc] {
		url := fmt.Sprintf("s3://%s/%s", m.id, m.items[itemKey].url)
		results[index] = mockStowItem{url: url, size: m.items[itemKey].size}
	}

	if endIndexExc == len(m.items) {
		cursor = ""
	} else {
		cursor = fmt.Sprintf("%d", endIndexExc)
	}
	return results, cursor, nil
}

func (m mockStowContainer) RemoveItem(id string) error {
	if _, found := m.items[id]; !found {
		return stow.ErrNotFound
	}

	delete(m.items, id)

	return nil
}

func (m *mockStowContainer) Put(name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error) {
	if m.putCB != nil {
		return m.putCB(name, r, size, metadata)
	}
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
	return io.NopCloser(bytes.NewReader([]byte{})), nil
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

func TestAwsBucketIsNotFound(t *testing.T) {
	for name, err := range errorWrappings(&smithy.GenericAPIError{Code: awsErrCodeNoSuchBucket, Message: "foo"}) {
		t.Run("detect is not found/"+name, func(t *testing.T) {
			assert.True(t, awsBucketIsNotFound(err))
		})
	}
	for name, err := range errorWrappings(&smithy.GenericAPIError{Code: "InvalidObjectState", Message: "foo"}) {
		t.Run("do not detect random errors/"+name, func(t *testing.T) {
			assert.False(t, awsBucketIsNotFound(err))
		})
	}
	t.Run("do not detect non aws errors", func(t *testing.T) {
		assert.False(t, awsBucketIsNotFound(errors2.New("foo")))
		assert.False(t, awsBucketIsNotFound(nil))
	})
}

func TestAwsBucketAlreadyExists(t *testing.T) {
	for name, err := range errorWrappings(&smithy.GenericAPIError{Code: awsErrCodeBucketAlreadyOwnedByYou, Message: "foo"}) {
		t.Run("detect already owned/"+name, func(t *testing.T) {
			assert.True(t, awsBucketAlreadyExists(err))
		})
	}
	for name, err := range errorWrappings(&os.PathError{Err: syscall.EEXIST}) {
		t.Run("detect file exists/"+name, func(t *testing.T) {
			assert.True(t, awsBucketAlreadyExists(err))
		})
	}
	for name, err := range errorWrappings(&smithy.GenericAPIError{Code: awsErrCodeNoSuchBucket, Message: "foo"}) {
		t.Run("do not detect random errors/"+name, func(t *testing.T) {
			assert.False(t, awsBucketAlreadyExists(err))
		})
	}
	t.Run("do not detect non aws errors", func(t *testing.T) {
		assert.False(t, awsBucketAlreadyExists(errors2.New("foo")))
		assert.False(t, awsBucketAlreadyExists(nil))
	})
}

func TestStowStore_CreateSignedURL(t *testing.T) {
	const container = "container"
	t.Run("Happy Path", func(t *testing.T) {
		fn := fQNFn["s3"]
		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
			CreateContainerCb: func(name string) (stow.Container, error) {
				if name == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, false, metrics)
		assert.NoError(t, err)

		actual, err := s.CreateSignedURL(context.TODO(), DataReference("https://container/path"), SignedURLProperties{})
		assert.NoError(t, err)
		assert.Equal(t, "path", actual.URL.String())
	})

	t.Run("Invalid URL", func(t *testing.T) {
		fn := fQNFn["s3"]
		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
			CreateContainerCb: func(name string) (stow.Container, error) {
				if name == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, false, metrics)
		assert.NoError(t, err)

		_, err = s.CreateSignedURL(context.TODO(), DataReference("://container/path"), SignedURLProperties{})
		assert.Error(t, err)
	})

	t.Run("Non existing container", func(t *testing.T) {
		fn := fQNFn["s3"]
		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
			CreateContainerCb: func(name string) (stow.Container, error) {
				if name == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, false, metrics)
		assert.NoError(t, err)

		_, err = s.CreateSignedURL(context.TODO(), DataReference("s3://container2/path"), SignedURLProperties{})
		assert.Error(t, err)
	})
}

func TestStowStore_ReadRaw(t *testing.T) {
	const container = "container"
	t.Run("Happy Path", func(t *testing.T) {
		ctx := context.Background()
		fn := fQNFn["s3"]
		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
			CreateContainerCb: func(name string) (stow.Container, error) {
				if name == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, false, metrics)
		assert.NoError(t, err)
		dataReference := writeTestFile(ctx, t, s, "s3://container/path")
		raw, err := s.ReadRaw(ctx, dataReference)
		assert.NoError(t, err)
		rawBytes, err := io.ReadAll(raw)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(rawBytes))
		assert.Equal(t, DataReference("s3://container"), s.GetBaseContainerFQN(context.TODO()))
	})

	t.Run("Exceeds limit", func(t *testing.T) {
		ctx := context.Background()
		fn := fQNFn["s3"]

		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
			CreateContainerCb: func(name string) (stow.Container, error) {
				if name == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, false, metrics)
		assert.NoError(t, err)
		dataReference := writeTestFileWithSize(ctx, t, s, "s3://container/path", 2*MiB+1)
		_, err = s.ReadRaw(ctx, dataReference)
		assert.Error(t, err)
		assert.True(t, IsExceedsLimit(err))
		assert.NotNil(t, errors.Cause(err))
	})

	t.Run("No Limit", func(t *testing.T) {
		ctx := context.Background()
		fn := fQNFn["s3"]
		GetConfig().Limits.GetLimitMegabytes = 0

		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
			CreateContainerCb: func(name string) (stow.Container, error) {
				if name == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, false, metrics)
		assert.NoError(t, err)
		dataReference := writeTestFileWithSize(ctx, t, s, "s3://container/path", 3*MiB)
		_, err = s.ReadRaw(ctx, dataReference)
		assert.Nil(t, err)
	})

	t.Run("Happy Path multi-container enabled", func(t *testing.T) {
		ctx := context.Background()
		fn := fQNFn["s3"]
		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				switch id {
				case container:
					return newMockStowContainer(container), nil
				case "bad-container":
					return newMockStowContainer("bad-container"), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
			CreateContainerCb: func(name string) (stow.Container, error) {
				if name == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, true, metrics)
		assert.NoError(t, err)
		dataReference := writeTestFile(ctx, t, s, "s3://bad-container/path")
		raw, err := s.ReadRaw(context.TODO(), dataReference)
		assert.NoError(t, err)
		rawBytes, err := io.ReadAll(raw)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(rawBytes))
		assert.Equal(t, DataReference("s3://container"), s.GetBaseContainerFQN(context.TODO()))
	})

	t.Run("Happy Path multi-container bad", func(t *testing.T) {
		fn := fQNFn["s3"]
		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
			CreateContainerCb: func(name string) (stow.Container, error) {
				if name == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, true, metrics)
		assert.NoError(t, err)
		err = s.WriteRaw(context.TODO(), "s3://bad-container/path", 0, Options{}, bytes.NewReader([]byte{}))
		assert.Error(t, err)
		_, err = s.Head(context.TODO(), "s3://bad-container/path")
		assert.Error(t, err)
		_, err = s.ReadRaw(context.TODO(), "s3://bad-container/path")
		assert.Error(t, err)
	})
}

func TestStowStore_List(t *testing.T) {
	const container = "container"
	t.Run("Listing", func(t *testing.T) {
		ctx := context.Background()
		fn := fQNFn["s3"]
		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
			CreateContainerCb: func(name string) (stow.Container, error) {
				if name == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, false, metrics)
		assert.NoError(t, err)
		writeTestFile(ctx, t, s, "s3://container/a/1")
		writeTestFile(ctx, t, s, "s3://container/a/2")
		var maxResults = 10
		var dataReference DataReference = "s3://container/a"
		items, cursor, err := s.List(ctx, dataReference, maxResults, NewCursorAtStart())
		assert.NoError(t, err)
		assert.Equal(t, NewCursorAtEnd(), cursor)
		assert.Equal(t, []DataReference{"s3://container/a/1", "s3://container/a/2"}, items)
	})

	t.Run("Listing with pagination", func(t *testing.T) {
		ctx := context.Background()
		fn := fQNFn["s3"]
		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
			CreateContainerCb: func(name string) (stow.Container, error) {
				if name == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, false, metrics)
		assert.NoError(t, err)
		writeTestFile(ctx, t, s, "s3://container/a/1")
		writeTestFile(ctx, t, s, "s3://container/a/2")
		var maxResults = 1
		var dataReference DataReference = "s3://container/a"
		items, cursor, err := s.List(ctx, dataReference, maxResults, NewCursorAtStart())
		assert.NoError(t, err)
		assert.Equal(t, []DataReference{"s3://container/a/1"}, items)
		items, _, err = s.List(ctx, dataReference, maxResults, cursor)
		assert.NoError(t, err)
		assert.Equal(t, []DataReference{"s3://container/a/2"}, items)
	})
}

func TestNewLocalStore(t *testing.T) {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
	t.Run("Valid config", func(t *testing.T) {
		store, err := newStowRawStore(context.TODO(), &Config{
			Stow: StowConfig{
				Kind: local.Kind,
				Config: map[string]string{
					local.ConfigKeyPath: "./",
				},
			},
			InitContainer: "testdata",
		}, metrics)

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
		_, err := newStowRawStore(context.TODO(), &Config{}, metrics)
		assert.Error(t, err)
	})

	t.Run("Initialize container", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "stdlib_local")
		assert.NoError(t, err)

		stats, err := os.Stat(tmpDir)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		store, err := newStowRawStore(context.TODO(), &Config{
			Stow: StowConfig{
				Kind: local.Kind,
				Config: map[string]string{
					local.ConfigKeyPath: tmpDir,
				},
			},
			InitContainer: "tmp",
		}, metrics)

		assert.NoError(t, err)
		assert.NotNil(t, store)

		stats, err = os.Stat(filepath.Join(tmpDir, "tmp"))
		assert.NoError(t, err)
		if assert.NotNil(t, stats) {
			assert.True(t, stats.IsDir())
		}
	})

	t.Run("missing init container", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "stdlib_local")
		assert.NoError(t, err)

		stats, err := os.Stat(tmpDir)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		store, err := newStowRawStore(context.TODO(), &Config{
			Stow: StowConfig{
				Kind: local.Kind,
				Config: map[string]string{
					local.ConfigKeyPath: tmpDir,
				},
			},
		}, metrics)

		assert.Error(t, err)
		assert.Nil(t, store)
	})

	t.Run("multi-container enabled", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "stdlib_local")
		assert.NoError(t, err)

		stats, err := os.Stat(tmpDir)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		store, err := newStowRawStore(context.TODO(), &Config{
			Stow: StowConfig{
				Kind: local.Kind,
				Config: map[string]string{
					local.ConfigKeyPath: tmpDir,
				},
			},
			InitContainer:         "tmp",
			MultiContainerEnabled: true,
		}, metrics)

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
	secretKey := "password"
	path := filepath.Join(t.TempDir(), "secret-key-path")
	err := os.WriteFile(path, []byte(secretKey), 0o600)
	assert.NoError(t, err)

	type args struct {
		cfg *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"fail", args{&Config{}}, true},
		{"google", args{&Config{
			InitContainer: "flyte",
			Stow: StowConfig{
				Kind: google.Kind,
				Config: map[string]string{
					google.ConfigProjectId: "x",
					google.ConfigScopes:    "y",
				},
			},
		}}, true},
		{"minio", args{&Config{
			Type:          TypeStow,
			InitContainer: "some-container",
			Stow: StowConfig{
				Kind: local.Kind,
				Config: map[string]string{
					"endpoint": "http://minio:9000",
				},
			},
		}}, true},
		{"secretKeyPath", args{&Config{
			Type:          TypeStow,
			InitContainer: "flyte",
			Stow: StowConfig{
				Kind: s3.Kind,
				Config: map[string]string{
					s3.ConfigAccessKeyID: "my-access-key-id",
					ConfigSecretKeyPath:  path,
				},
			},
		}}, false},
		{"secretKeyPath not found", args{&Config{
			Type:          TypeStow,
			InitContainer: "flyte",
			Stow: StowConfig{
				Kind: s3.Kind,
				Config: map[string]string{
					s3.ConfigAccessKeyID: "my-access-key-id",
					ConfigSecretKeyPath:  filepath.Join(t.TempDir(), "wrong"),
				},
			},
		}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newStowRawStore(context.TODO(), tt.args.cfg, metrics)
			if tt.wantErr {
				require.Error(t, err, "newStowRawStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			require.NoError(t, err, "newStowRawStore() error = %v, wantErr %v", err, tt.wantErr)
			require.NotNil(t, got, "Expected rawstore, found nil!")
		})
	}
}

func TestLoadContainer(t *testing.T) {
	container := "container"
	t.Run("Create if not found", func(t *testing.T) {
		stowStore := StowStore{
			loc: &mockStowLoc{
				ContainerCb: func(id string) (stow.Container, error) {
					if id == container {
						return newMockStowContainer(container), stow.ErrNotFound
					}
					return nil, fmt.Errorf("container is not supported")
				},
				CreateContainerCb: func(name string) (stow.Container, error) {
					if name == container {
						return newMockStowContainer(container), nil
					}
					return nil, fmt.Errorf("container is not supported")
				},
			},
		}
		stowContainer, err := stowStore.LoadContainer(context.Background(), "container", true)
		assert.NoError(t, err)
		assert.Equal(t, container, stowContainer.ID())
	})
	t.Run("Create if not found with error", func(t *testing.T) {
		stowStore := StowStore{
			loc: &mockStowLoc{
				ContainerCb: func(id string) (stow.Container, error) {
					return nil, stow.ErrNotFound
				},
				CreateContainerCb: func(name string) (stow.Container, error) {
					if name == container {
						return nil, fmt.Errorf("foo")
					}
					return nil, fmt.Errorf("container is not supported")
				},
			},
		}
		_, err := stowStore.LoadContainer(context.TODO(), "container", true)
		assert.EqualError(t, err, "unable to initialize container [container]. Error: foo")
	})
	t.Run("No create if not found", func(t *testing.T) {
		stowStore := StowStore{
			loc: &mockStowLoc{
				ContainerCb: func(id string) (stow.Container, error) {
					if id == container {
						return newMockStowContainer(container), stow.ErrNotFound
					}
					return nil, fmt.Errorf("container is not supported")
				},
			},
		}
		_, err := stowStore.LoadContainer(context.TODO(), "container", false)
		assert.EqualError(t, err, stow.ErrNotFound.Error())
	})
}

func TestStowStore_WriteRaw(t *testing.T) {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
	const container = "container"
	fn := fQNFn["s3"]
	// newStore returns a store whose configured container fails every Put with putErr, and the
	// container CreateContainer hands out.
	newStore := func(t *testing.T, putErr error) (*StowStore, *mockStowContainer, *bool) {
		created := newMockStowContainer(container)
		createCalled := false
		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					mockStowContainer := newMockStowContainer(container)
					mockStowContainer.putCB = func(string, io.Reader, int64, map[string]interface{}) (stow.Item, error) {
						return nil, putErr
					}
					return mockStowContainer, nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
			CreateContainerCb: func(name string) (stow.Container, error) {
				createCalled = true
				if name == container {
					return created, nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, true, metrics)
		assert.NoError(t, err)
		return s, created, &createCalled
	}
	noSuchBucket := &smithy.GenericAPIError{Code: awsErrCodeNoSuchBucket, Message: "foo"}

	t.Run("create container when not found and write again", func(t *testing.T) {
		s, created, createCalled := newStore(t, noSuchBucket)
		err := s.WriteRaw(t.Context(), DataReference("s3://container/path"), 5, Options{}, bytes.NewReader([]byte("hello")))
		assert.NoError(t, err)
		assert.True(t, *createCalled)
		// The object was written to the created container, which is now the one in use.
		assert.Contains(t, created.items, "path")
		stored, ok := s.dynamicContainerMap.Load(locationIDMain.String() + container)
		assert.True(t, ok)
		assert.Same(t, created, stored)
	})
	t.Run("create container when not found, stow wraps with %w", func(t *testing.T) {
		s, created, createCalled := newStore(t, fmt.Errorf("PutObject, putting object: %w", noSuchBucket))
		err := s.WriteRaw(t.Context(), DataReference("s3://container/path"), 5, Options{}, bytes.NewReader([]byte("hello")))
		assert.NoError(t, err)
		assert.True(t, *createCalled)
		assert.Contains(t, created.items, "path")
	})
	t.Run("container created concurrently is loaded and written to", func(t *testing.T) {
		existing := newMockStowContainer(container)
		puts := 0
		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id != container {
					return nil, fmt.Errorf("container is not supported")
				}
				// The first lookup is the configured container, whose bucket is gone. The next
				// one is the reload after another writer created the bucket.
				if puts == 0 {
					missing := newMockStowContainer(container)
					missing.putCB = func(string, io.Reader, int64, map[string]interface{}) (stow.Item, error) {
						puts++
						return nil, noSuchBucket
					}
					return missing, nil
				}
				return existing, nil
			},
			CreateContainerCb: func(string) (stow.Container, error) {
				return nil, &smithy.GenericAPIError{Code: awsErrCodeBucketAlreadyOwnedByYou, Message: "foo"}
			},
		}, nil, true, metrics)
		assert.NoError(t, err)

		err = s.WriteRaw(t.Context(), DataReference("s3://container/path"), 5, Options{}, bytes.NewReader([]byte("hello")))
		assert.NoError(t, err)
		assert.Contains(t, existing.items, "path")
	})
	t.Run("write again continues from where the reader started", func(t *testing.T) {
		s, created, _ := newStore(t, noSuchBucket)
		var written []byte
		created.putCB = func(_ string, r io.Reader, _ int64, _ map[string]interface{}) (stow.Item, error) {
			var err error
			written, err = io.ReadAll(r)
			return mockStowItem{}, err
		}
		reader := bytes.NewReader([]byte("skip hello"))
		_, err := reader.Seek(5, io.SeekStart)
		assert.NoError(t, err)
		assert.NoError(t, s.WriteRaw(t.Context(), DataReference("s3://container/path"), 5, Options{}, reader))
		assert.Equal(t, "hello", string(written))
	})
	t.Run("data that cannot be read again is an error, not a silent loss", func(t *testing.T) {
		s, created, createCalled := newStore(t, noSuchBucket)
		raw := io.LimitReader(bytes.NewReader([]byte("hello")), 5)
		err := s.WriteRaw(t.Context(), DataReference("s3://container/path"), 5, Options{}, raw)
		assert.ErrorContains(t, err, "cannot be read again")
		assert.True(t, *createCalled)
		assert.Empty(t, created.items)
	})
	t.Run("bubble up generic put errors", func(t *testing.T) {
		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					mockStowContainer := newMockStowContainer(container)
					mockStowContainer.putCB = func(name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error) {
						return nil, errors2.New("foo")
					}
					return mockStowContainer, nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, true, metrics)
		assert.NoError(t, err)
		err = s.WriteRaw(context.TODO(), DataReference("s3://container/path"), 0, Options{}, bytes.NewReader([]byte{}))
		assert.EqualError(t, err, "Failed to write data [0b] to path [path].: foo")
	})
}

func TestStowStore_fQNFn(t *testing.T) {
	assert.Equal(t, DataReference("s3://bucket"), fQNFn[s3.Kind]("bucket"))
	assert.Equal(t, DataReference("gs://bucket"), fQNFn[google.Kind]("bucket"))
	assert.Equal(t, DataReference("os://bucket"), fQNFn[oracle.Kind]("bucket"))
	assert.Equal(t, DataReference("sw://bucket"), fQNFn[swift.Kind]("bucket"))
	assert.Equal(t, DataReference("abfs://bucket"), fQNFn[azure.Kind]("bucket"))
	assert.Equal(t, DataReference("file://bucket"), fQNFn[local.Kind]("bucket"))
}

func TestPrimarySchemeForConfig(t *testing.T) {
	t.Run("built-in stow kind", func(t *testing.T) {
		scheme, err := primarySchemeForConfig(&Config{Type: TypeStow, Stow: StowConfig{Kind: google.Kind}})
		assert.NoError(t, err)
		assert.Equal(t, "gs", scheme)
	})

	t.Run("custom kind derives scheme from registered fQNFn", func(t *testing.T) {
		const kind = "test-custom-primary-kind"
		assert.NoError(t, RegisterStowKind(kind, func(bucket string) DataReference {
			return DataReference("tcpk://" + bucket)
		}))
		// kindToScheme has no entry for this kind, so the scheme must be derived from fQNFn — this is
		// what lets an out-of-tree RegisterStowKind backend serve as the primary store.
		scheme, err := primarySchemeForConfig(&Config{Type: TypeStow, Stow: StowConfig{Kind: kind}})
		assert.NoError(t, err)
		assert.Equal(t, "tcpk", scheme)
	})

	t.Run("unknown kind errors", func(t *testing.T) {
		_, err := primarySchemeForConfig(&Config{Type: TypeStow, Stow: StowConfig{Kind: "no-such-kind"}})
		assert.Error(t, err)
	})
}

func TestStowFactory_AmbientDialForUnconfiguredScheme(t *testing.T) {
	t.Run("unconfigured scheme dials with ambient credentials", func(t *testing.T) {
		// No Schemes entry for s3: the factory must derive the stow kind from the scheme and dial with
		// ambient credentials (no explicit access key/secret — the provider's default credential chain).
		// Stub the dial so the test stays hermetic (the real stow S3 driver would depend on the ambient
		// AWS region/credential chain); assert on the kind and config the factory resolved instead.
		var gotKind string
		var gotCfg stow.ConfigMap
		orig := stowDial
		stowDial = func(_ *http.Client, kind string, cfgMap stow.ConfigMap) (stow.Location, error) {
			gotKind, gotCfg = kind, cfgMap
			return nil, nil // a secondary scheme has an empty base container, so loc is never dereferenced
		}
		defer func() { stowDial = orig }()

		store, err := stowFactory(context.TODO(), "s3", "s3://bucket/key", &Config{}, nil, metrics)
		assert.NoError(t, err)
		assert.IsType(t, &StowStore{}, store)
		assert.Equal(t, s3.Kind, gotKind)
		assert.NotContains(t, gotCfg, s3.ConfigAccessKeyID, "ambient dial must not inject explicit credentials")
		assert.NotContains(t, gotCfg, s3.ConfigSecretKey, "ambient dial must not inject explicit credentials")
	})

	t.Run("scheme with no registered stow kind errors", func(t *testing.T) {
		_, err := stowFactory(context.TODO(), "not-a-scheme", "not-a-scheme://b/k", &Config{}, nil, metrics)
		assert.Error(t, err)
	})

	t.Run("local backend without a path fails fast", func(t *testing.T) {
		// The local (file://) backend can't be dialed with ambient config; it needs an explicit root
		// path, so an unconfigured file:// scheme must error deterministically with an actionable message.
		_, err := stowFactory(context.TODO(), "file", "file://root/key", &Config{}, nil, metrics)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), local.ConfigKeyPath)
	})
}

func TestStowStore_Delete(t *testing.T) {
	const container = "container"

	t.Run("Happy Path", func(t *testing.T) {
		ctx := context.TODO()
		fn := fQNFn["s3"]

		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
			CreateContainerCb: func(name string) (stow.Container, error) {
				if name == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, false, metrics)
		assert.NoError(t, err)

		dataReference := writeTestFile(ctx, t, s, "s3://container/path")

		err = s.Delete(ctx, dataReference)
		assert.NoError(t, err)

		metadata, err := s.Head(ctx, dataReference)
		assert.NoError(t, err)
		assert.False(t, metadata.Exists())
	})

	t.Run("Happy Path multi-container enabled", func(t *testing.T) {
		ctx := context.TODO()
		fn := fQNFn["s3"]

		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				switch id {
				case container:
					return newMockStowContainer(container), nil
				case "bad-container":
					return newMockStowContainer("bad-container"), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
			CreateContainerCb: func(name string) (stow.Container, error) {
				if name == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, true, metrics)
		assert.NoError(t, err)

		dataReference := writeTestFile(ctx, t, s, "s3://container/path")
		dataReference2 := writeTestFile(ctx, t, s, "s3://bad-container/path")

		err = s.Delete(ctx, dataReference)
		assert.NoError(t, err)
		err = s.Delete(ctx, dataReference2)
		assert.NoError(t, err)

		metadata, err := s.Head(ctx, dataReference)
		assert.NoError(t, err)
		assert.False(t, metadata.Exists())
		metadata, err = s.Head(ctx, dataReference2)
		assert.NoError(t, err)
		assert.False(t, metadata.Exists())
	})

	t.Run("Unknown item", func(t *testing.T) {
		ctx := context.TODO()
		fn := fQNFn["s3"]

		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
			CreateContainerCb: func(name string) (stow.Container, error) {
				if name == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, false, metrics)
		assert.NoError(t, err)

		dataReference := writeTestFile(ctx, t, s, "s3://container/path")

		err = s.Delete(ctx, DataReference("s3://container/bad-path"))
		assert.Error(t, err)
		assert.True(t, errors.Is(err, stow.ErrNotFound))

		metadata, err := s.Head(ctx, dataReference)
		assert.NoError(t, err)
		assert.True(t, metadata.Exists())
	})

	t.Run("Unknown container", func(t *testing.T) {
		ctx := context.TODO()
		fn := fQNFn["s3"]

		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
			CreateContainerCb: func(name string) (stow.Container, error) {
				if name == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, false, metrics)
		assert.NoError(t, err)

		dataReference := writeTestFile(ctx, t, s, "s3://container/path")

		err = s.Delete(ctx, DataReference("s3://bad-container/path"))
		assert.Error(t, err)
		assert.True(t, errors.Is(err, stow.ErrNotFound))

		metadata, err := s.Head(ctx, dataReference)
		assert.NoError(t, err)
		assert.True(t, metadata.Exists())
	})

	t.Run("Invalid data reference", func(t *testing.T) {
		ctx := context.TODO()
		fn := fQNFn["s3"]

		s, err := NewStowRawStore(fn(container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
			CreateContainerCb: func(name string) (stow.Container, error) {
				if name == container {
					return newMockStowContainer(container), nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, false, metrics)
		assert.NoError(t, err)

		err = s.Delete(ctx, DataReference("://bad-container/path"))
		assert.Error(t, err)
	})
}

func writeTestFile(ctx context.Context, t *testing.T, s *StowStore, path string) DataReference {
	return writeTestFileWithSize(ctx, t, s, path, 0)
}

func writeTestFileWithSize(ctx context.Context, t *testing.T, s *StowStore, path string, size int64) DataReference {
	reference := DataReference(path)

	err := s.WriteRaw(ctx, reference, size, Options{}, bytes.NewReader([]byte{}))
	assert.NoError(t, err)

	metadata, err := s.Head(ctx, reference)
	assert.NoError(t, err)
	assert.True(t, metadata.Exists())

	return reference
}

// mockStowCopierContainer is a container that also copies on the server side.
type mockStowCopierContainer struct {
	*mockStowContainer
	copyCB func(ctx context.Context, src stow.Item, name string) (stow.Item, error)
}

func (m *mockStowCopierContainer) Copy(ctx context.Context, src stow.Item, name string) (stow.Item, error) {
	return m.copyCB(ctx, src, name)
}

func TestStowStore_CopyRaw(t *testing.T) {
	const container = "container"
	const source = DataReference("s3://container/src/outputs.pb")
	const destination = DataReference("s3://container/dst/outputs.pb")

	newStore := func(t *testing.T, c stow.Container) *StowStore {
		s, err := NewStowRawStore(fQNFn["s3"](container), &mockStowLoc{
			ContainerCb: func(id string) (stow.Container, error) {
				if id == container {
					return c, nil
				}
				return nil, fmt.Errorf("container is not supported")
			},
		}, nil, false, metrics)
		assert.NoError(t, err)
		return s
	}

	// newContainer returns a container holding the source item, which is over the download limit:
	// no copy may read it through ReadRaw. Any Put fails the test unless a streamed copy is expected.
	newContainer := func(t *testing.T, putAllowed bool) (*mockStowContainer, *int) {
		puts := 0
		c := newMockStowContainer(container)
		c.items["src/outputs.pb"] = mockStowItem{url: "src/outputs.pb", size: 1 << 30}
		c.putCB = func(name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error) {
			puts++
			if !putAllowed {
				t.Errorf("unexpected Put of [%v]", name)
			}
			// The reader is the opened source item, handed over as is: nothing was buffered.
			_, buffered := r.(*bytes.Reader)
			assert.False(t, buffered, "source was buffered before the upload")
			assert.Equal(t, "dst/outputs.pb", name)
			assert.Equal(t, int64(1<<30), size)
			return mockStowItem{url: name, size: size}, nil
		}
		return c, &puts
	}

	t.Run("server side", func(t *testing.T) {
		base, _ := newContainer(t, false)
		var copied []string
		s := newStore(t, &mockStowCopierContainer{
			mockStowContainer: base,
			copyCB: func(_ context.Context, src stow.Item, name string) (stow.Item, error) {
				copied = append(copied, src.ID()+" -> "+name)
				return mockStowItem{url: name}, nil
			},
		})

		assert.NoError(t, s.CopyRaw(context.Background(), source, destination, Options{}))
		assert.Equal(t, []string{"src/outputs.pb -> dst/outputs.pb"}, copied)
	})

	t.Run("source not found", func(t *testing.T) {
		base, _ := newContainer(t, false)
		s := newStore(t, &mockStowCopierContainer{
			mockStowContainer: base,
			copyCB: func(context.Context, stow.Item, string) (stow.Item, error) {
				t.Error("unexpected Copy")
				return nil, nil
			},
		})

		err := s.CopyRaw(context.Background(), "s3://container/missing/outputs.pb", destination, Options{})
		assert.True(t, IsNotFound(err), "got %v", err)
	})

	t.Run("copy fails", func(t *testing.T) {
		base, _ := newContainer(t, false)
		s := newStore(t, &mockStowCopierContainer{
			mockStowContainer: base,
			copyCB: func(context.Context, stow.Item, string) (stow.Item, error) {
				return nil, fmt.Errorf("access denied")
			},
		})

		assert.ErrorContains(t, s.CopyRaw(context.Background(), source, destination, Options{}), "access denied")
	})

	t.Run("container without copy streams", func(t *testing.T) {
		base, puts := newContainer(t, true)
		s := newStore(t, base)

		assert.NoError(t, s.CopyRaw(context.Background(), source, destination, Options{}))
		assert.Equal(t, 1, *puts)
	})
}

// contextRecorder keeps the contexts the stow methods with a context are called with.
type contextRecorder struct {
	got map[string]context.Context
}

func (r *contextRecorder) record(method string, ctx context.Context) {
	r.got[method] = ctx
}

// contextStowLoc, contextStowContainer and contextStowItem have the stow methods with a context
// on top of the mocks without one.
type contextStowLoc struct {
	mockStowLoc
	*contextRecorder
}

func (l contextStowLoc) ContainerContext(ctx context.Context, id string) (stow.Container, error) {
	l.record("Container", ctx)
	return l.Container(id)
}

func (l contextStowLoc) CreateContainerContext(ctx context.Context, name string) (stow.Container, error) {
	l.record("CreateContainer", ctx)
	return l.CreateContainer(name)
}

func (l contextStowLoc) ContainersContext(context.Context, string, string, int) ([]stow.Container, string, error) {
	return nil, "", fmt.Errorf("not implemented")
}

func (l contextStowLoc) RemoveContainerContext(context.Context, string) error {
	return fmt.Errorf("not implemented")
}

func (l contextStowLoc) ItemByURLContext(context.Context, *url.URL) (stow.Item, error) {
	return nil, fmt.Errorf("not implemented")
}

type contextStowContainer struct {
	*mockStowContainer
	*contextRecorder
}

func (c contextStowContainer) ItemContext(ctx context.Context, id string) (stow.Item, error) {
	c.record("Item", ctx)
	item, err := c.Item(id)
	if err != nil {
		return nil, err
	}
	return contextStowItem{Item: item, contextRecorder: c.contextRecorder}, nil
}

func (c contextStowContainer) ItemsContext(ctx context.Context, prefix, cursor string, count int) ([]stow.Item, string, error) {
	c.record("Items", ctx)
	return c.Items(prefix, cursor, count)
}

func (c contextStowContainer) RemoveItemContext(ctx context.Context, id string) error {
	c.record("RemoveItem", ctx)
	return c.RemoveItem(id)
}

func (c contextStowContainer) PutContext(ctx context.Context, name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error) {
	c.record("Put", ctx)
	return c.Put(name, r, size, metadata)
}

type contextStowItem struct {
	stow.Item
	*contextRecorder
}

func (i contextStowItem) OpenContext(ctx context.Context) (io.ReadCloser, error) {
	i.record("Open", ctx)
	return i.Open()
}

func (i contextStowItem) ETagContext(ctx context.Context) (string, error) {
	i.record("ETag", ctx)
	return i.ETag()
}

func (i contextStowItem) LastModContext(ctx context.Context) (time.Time, error) {
	i.record("LastMod", ctx)
	return i.LastMod()
}

func (i contextStowItem) MetadataContext(ctx context.Context) (map[string]interface{}, error) {
	i.record("Metadata", ctx)
	return i.Metadata()
}

type contextTestKey struct{}

func TestStowStore_PassesContext(t *testing.T) {
	const container = "container"
	ref := DataReference("s3://container/path")
	ctx := context.WithValue(t.Context(), contextTestKey{}, "value")

	newStore := func(t *testing.T) (*StowStore, *contextRecorder) {
		recorder := &contextRecorder{got: map[string]context.Context{}}
		c := contextStowContainer{mockStowContainer: newMockStowContainer(container), contextRecorder: recorder}
		s, err := NewStowRawStore(fQNFn["s3"](container), contextStowLoc{
			contextRecorder: recorder,
			mockStowLoc: mockStowLoc{
				ContainerCb:       func(string) (stow.Container, error) { return c, nil },
				CreateContainerCb: func(string) (stow.Container, error) { return c, nil },
			},
		}, nil, true, metrics)
		require.NoError(t, err)
		// forget the contexts of the store's construction
		clear(recorder.got)
		return s, recorder
	}

	// assertPassed checks that each of the stow methods was called with ctx and no other one was.
	assertPassed := func(t *testing.T, recorder *contextRecorder, methods ...string) {
		t.Helper()
		assert.Len(t, recorder.got, len(methods))
		for _, method := range methods {
			assert.Equal(t, ctx, recorder.got[method], method)
		}
	}

	write := func(t *testing.T, s *StowStore) {
		t.Helper()
		require.NoError(t, s.WriteRaw(ctx, ref, 0, Options{}, bytes.NewReader([]byte{})))
	}

	t.Run("WriteRaw", func(t *testing.T) {
		s, recorder := newStore(t)
		write(t, s)
		assertPassed(t, recorder, "Put")
	})

	t.Run("Head", func(t *testing.T) {
		s, recorder := newStore(t)
		write(t, s)
		metadata, err := s.Head(ctx, ref)
		require.NoError(t, err)
		assert.True(t, metadata.Exists())
		assertPassed(t, recorder, "Put", "Item", "Metadata", "ETag")
	})

	t.Run("ReadRaw", func(t *testing.T) {
		s, recorder := newStore(t)
		write(t, s)
		r, err := s.ReadRaw(ctx, ref)
		require.NoError(t, err)
		require.NoError(t, r.Close())
		assertPassed(t, recorder, "Put", "Item", "Open")
	})

	t.Run("List", func(t *testing.T) {
		s, recorder := newStore(t)
		_, _, err := s.List(ctx, ref, 10, NewCursorAtStart())
		require.NoError(t, err)
		assertPassed(t, recorder, "Items")
	})

	t.Run("Delete", func(t *testing.T) {
		s, recorder := newStore(t)
		write(t, s)
		require.NoError(t, s.Delete(ctx, ref))
		assertPassed(t, recorder, "Put", "RemoveItem")
	})

	t.Run("CopyRaw", func(t *testing.T) {
		s, recorder := newStore(t)
		write(t, s)
		require.NoError(t, s.CopyRaw(ctx, ref, DataReference("s3://container/copy"), Options{}))
		// the container of the test is no stow.Copier, so the copy is streamed
		assertPassed(t, recorder, "Put", "Item", "Metadata", "Open")
	})

	t.Run("LoadContainer", func(t *testing.T) {
		s, recorder := newStore(t)
		_, err := s.LoadContainer(ctx, "other", false)
		require.NoError(t, err)
		assertPassed(t, recorder, "Container")
	})

	t.Run("LoadContainer creates a missing container", func(t *testing.T) {
		recorder := &contextRecorder{got: map[string]context.Context{}}
		s := &StowStore{loc: contextStowLoc{
			contextRecorder: recorder,
			mockStowLoc: mockStowLoc{
				ContainerCb:       func(string) (stow.Container, error) { return nil, stow.ErrNotFound },
				CreateContainerCb: func(string) (stow.Container, error) { return newMockStowContainer(container), nil },
			},
		}}
		_, err := s.LoadContainer(ctx, container, true)
		require.NoError(t, err)
		assertPassed(t, recorder, "Container", "CreateContainer")
	})

	t.Run("a cancelled context stops a store without context methods", func(t *testing.T) {
		c := newMockStowContainer(container)
		s, err := NewStowRawStore(fQNFn["s3"](container), &mockStowLoc{
			ContainerCb:       func(string) (stow.Container, error) { return c, nil },
			CreateContainerCb: func(string) (stow.Container, error) { return c, nil },
		}, nil, false, metrics)
		require.NoError(t, err)

		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		assert.ErrorIs(t, s.WriteRaw(cancelled, ref, 0, Options{}, bytes.NewReader([]byte{})), context.Canceled)
		_, err = s.Head(cancelled, ref)
		assert.ErrorIs(t, err, context.Canceled)
		_, err = s.ReadRaw(cancelled, ref)
		assert.ErrorIs(t, err, context.Canceled)
		_, _, err = s.List(cancelled, ref, 10, NewCursorAtStart())
		assert.ErrorIs(t, err, context.Canceled)
		assert.ErrorIs(t, s.Delete(cancelled, ref), context.Canceled)
		assert.Empty(t, c.items)
	})
}
