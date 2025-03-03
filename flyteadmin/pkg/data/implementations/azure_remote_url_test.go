package implementations

import (
	"context"
	commonMocks "github.com/flyteorg/flyte/flyteadmin/pkg/common/mocks"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockMetadata struct{}

func (m mockMetadata) ContentMD5() string {
	return ""
}

func (m mockMetadata) Exists() bool {
	return true
}

func (m mockMetadata) Size() int64 {
	return 1
}

func (m mockMetadata) Etag() string {
	return "etag"
}

func TestAzureGet(t *testing.T) {
	inputUri := "abfs//test/data"
	mockStorage := commonMocks.GetMockStorageClient()
	mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).HeadCb =
		func(ctx context.Context, reference storage.DataReference) (storage.Metadata, error) {
			return mockMetadata{}, nil
		}
	remoteURL := AzureRemoteURL{
		remoteDataStoreClient: *mockStorage, presignDuration: 1,
	}

	result, _ := remoteURL.Get(context.TODO(), inputUri)
	assert.Contains(t, inputUri, result.Url)
}
