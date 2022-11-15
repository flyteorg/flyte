package dataproxy

import (
	"context"
	"testing"
	"time"

	commonMocks "github.com/flyteorg/flyteadmin/pkg/common/mocks"
	stdlibConfig "github.com/flyteorg/flytestdlib/config"

	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	"github.com/flyteorg/flyteadmin/pkg/config"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)
	s, err := NewService(config.DataProxyConfig{
		Upload: config.DataProxyUploadConfig{},
	}, dataStore)
	assert.NoError(t, err)
	assert.NotNil(t, s)
}

func init() {
	labeled.SetMetricKeys(contextutils.DomainKey)
}

func Test_createStorageLocation(t *testing.T) {
	dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)
	loc, err := createStorageLocation(context.Background(), dataStore, config.DataProxyUploadConfig{
		StoragePrefix: "blah",
	})
	assert.NoError(t, err)
	assert.Equal(t, "/blah", loc.String())
}

func TestCreateUploadLocation(t *testing.T) {
	dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)
	s, err := NewService(config.DataProxyConfig{}, dataStore)
	assert.NoError(t, err)
	t.Run("No project/domain", func(t *testing.T) {
		_, err = s.CreateUploadLocation(context.Background(), &service.CreateUploadLocationRequest{})
		assert.Error(t, err)
	})

	t.Run("unsupported operation by InMemory DataStore", func(t *testing.T) {
		_, err = s.CreateUploadLocation(context.Background(), &service.CreateUploadLocationRequest{
			Project: "hello",
			Domain:  "world",
		})
		assert.Error(t, err)
	})

	t.Run("Invalid expiry", func(t *testing.T) {
		_, err = s.CreateUploadLocation(context.Background(), &service.CreateUploadLocationRequest{
			Project:   "hello",
			Domain:    "world",
			ExpiresIn: durationpb.New(-time.Hour),
		})
		assert.Error(t, err)
	})
}

func TestCreateDownloadLocation(t *testing.T) {
	dataStore := commonMocks.GetMockStorageClient()
	s, err := NewService(config.DataProxyConfig{Download: config.DataProxyDownloadConfig{MaxExpiresIn: stdlibConfig.Duration{Duration: time.Hour}}}, dataStore)
	assert.NoError(t, err)

	t.Run("Invalid expiry", func(t *testing.T) {
		_, err = s.CreateDownloadLocation(context.Background(), &service.CreateDownloadLocationRequest{
			NativeUrl: "s3://bucket/key",
			ExpiresIn: durationpb.New(-time.Hour),
		})
		assert.Error(t, err)
	})

	t.Run("valid config", func(t *testing.T) {
		_, err = s.CreateDownloadLocation(context.Background(), &service.CreateDownloadLocationRequest{
			NativeUrl: "s3://bucket/key",
			ExpiresIn: durationpb.New(time.Hour),
		})
		assert.NoError(t, err)
	})

	t.Run("use default ExpiresIn", func(t *testing.T) {
		_, err = s.CreateDownloadLocation(context.Background(), &service.CreateDownloadLocationRequest{
			NativeUrl: "s3://bucket/key",
		})
		assert.NoError(t, err)
	})

	t.Run("invalid URL", func(t *testing.T) {
		_, err = s.CreateDownloadLocation(context.Background(), &service.CreateDownloadLocationRequest{
			NativeUrl: "bucket/key",
		})
		assert.NoError(t, err)
	})
}
