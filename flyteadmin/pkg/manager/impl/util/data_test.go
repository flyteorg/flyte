package util

import (
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

func TestShouldFetchData(t *testing.T) {
	t.Run("local config", func(t *testing.T) {
		assert.True(t, ShouldFetchData(&interfaces.RemoteDataConfig{
			Scheme:         common.Local,
			MaxSizeInBytes: 100,
		}, admin.UrlBlob{
			Bytes: 200,
		}))
	})
	t.Run("no config", func(t *testing.T) {
		assert.True(t, ShouldFetchData(&interfaces.RemoteDataConfig{
			Scheme:         common.None,
			MaxSizeInBytes: 100,
		}, admin.UrlBlob{
			Bytes: 200,
		}))
	})
	t.Run("max size under limit", func(t *testing.T) {
		assert.True(t, ShouldFetchData(&interfaces.RemoteDataConfig{
			Scheme:         common.AWS,
			MaxSizeInBytes: 1000,
		}, admin.UrlBlob{
			Bytes: 200,
		}))
	})
	t.Run("max size over limit", func(t *testing.T) {
		assert.False(t, ShouldFetchData(&interfaces.RemoteDataConfig{
			Scheme:         common.AWS,
			MaxSizeInBytes: 100,
		}, admin.UrlBlob{
			Bytes: 200,
		}))
	})
}

func TestShouldFetchOutputData(t *testing.T) {
	t.Run("local config", func(t *testing.T) {
		assert.True(t, ShouldFetchOutputData(&interfaces.RemoteDataConfig{
			Scheme:         common.Local,
			MaxSizeInBytes: 100,
		}, admin.UrlBlob{
			Bytes: 200,
		}, "s3://foo/bar.txt"))
	})
	t.Run("max size under limit", func(t *testing.T) {
		assert.True(t, ShouldFetchOutputData(&interfaces.RemoteDataConfig{
			Scheme:         common.AWS,
			MaxSizeInBytes: 1000,
		}, admin.UrlBlob{
			Bytes: 200,
		}, "s3://foo/bar.txt"))
	})
	t.Run("output uri empty", func(t *testing.T) {
		assert.False(t, ShouldFetchOutputData(&interfaces.RemoteDataConfig{
			Scheme:         common.AWS,
			MaxSizeInBytes: 1000,
		}, admin.UrlBlob{
			Bytes: 200,
		}, ""))
	})
}
