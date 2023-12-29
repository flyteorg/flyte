package implementations

import (
	"context"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/mocks"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/stretchr/testify/assert"
)

const noopFileSize = int64(1256)

type MockMetadata struct{}

func (m MockMetadata) Exists() bool {
	return true
}

func (m MockMetadata) Size() int64 {
	return noopFileSize
}

func (m MockMetadata) Etag() string {
	return "etag"
}

func getMockStorage() common.DatastoreClient {
	mockStorage := &mocks.DatastoreClient{}
	mockStorage.OnHeadMatch(mock.Anything, mock.Anything).Return(MockMetadata{}, nil)
	return mockStorage
}

func TestNoopRemoteURLGet(t *testing.T) {
	noopRemoteURL := NewNoopRemoteURL(getMockStorage())
	urlBlob, err := noopRemoteURL.Get(context.Background(), "uri")
	assert.Nil(t, err)
	assert.NotEmpty(t, urlBlob)
	assert.Equal(t, "uri", urlBlob.Url)
	assert.Equal(t, noopFileSize, urlBlob.Bytes)
}
