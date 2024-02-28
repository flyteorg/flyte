package cacheservice

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	catalogmocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog/mocks"
	mocks2 "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
)

func TestFallBack_Get(t *testing.T) {
	ctx := context.Background()

	mockEntry := catalog.NewCatalogEntry(&mocks2.OutputReader{}, catalog.NewStatus(core.CatalogCacheStatus_CACHE_POPULATED, &core.CatalogMetadata{}))

	testCases := []struct {
		name                       string
		mockCacheServiceGetError   error
		mockCatalogClientGetError  error
		mockCatalogClientGetEntry  catalog.Entry
		mockCacheServicePurError   error
		expectCatalogClientGetCall bool
		expectCacheServicePutCall  bool
		expectError                bool
		expectedClientErrCode      codes.Code
	}{
		{
			name:                       "cache service get error",
			mockCacheServiceGetError:   status.Error(codes.Internal, "get error"),
			expectCatalogClientGetCall: false,
			expectCacheServicePutCall:  false,
			expectError:                true,
		},
		{
			name:                       "cache service found",
			mockCacheServiceGetError:   nil,
			expectCatalogClientGetCall: false,
			expectCacheServicePutCall:  false,
			expectError:                false,
		},
		{
			name:                       "cache service not found, catalog client get error",
			mockCacheServiceGetError:   status.Error(codes.NotFound, "not found"),
			mockCatalogClientGetError:  status.Error(codes.Internal, "get error"),
			expectCatalogClientGetCall: true,
			expectCacheServicePutCall:  false,
			expectError:                true,
		},
		{
			name:                       "cache service not found, catalog client not found",
			mockCacheServiceGetError:   status.Error(codes.NotFound, "not found"),
			mockCatalogClientGetError:  status.Error(codes.NotFound, "get error"),
			expectCatalogClientGetCall: true,
			expectCacheServicePutCall:  false,
			expectError:                true,
		},
		{
			name:                       "cache service not found, catalog client found",
			mockCacheServiceGetError:   status.Error(codes.NotFound, "not found"),
			mockCatalogClientGetError:  nil,
			mockCatalogClientGetEntry:  mockEntry,
			mockCacheServicePurError:   nil,
			expectCatalogClientGetCall: true,
			expectCacheServicePutCall:  true,
			expectError:                false,
		},
		{
			name:                       "cache client put error",
			mockCacheServiceGetError:   status.Error(codes.NotFound, "not found"),
			mockCatalogClientGetError:  nil,
			mockCacheServicePurError:   status.Error(codes.Internal, "put error"),
			expectCatalogClientGetCall: true,
			expectCacheServicePutCall:  true,
			expectError:                false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCacheClient := &catalogmocks.Client{}
			mockCacheClient.On("Get",
				ctx,
				mock.Anything,
			).Return(catalog.Entry{}, tc.mockCacheServiceGetError)
			mockCacheClient.On("Put",
				ctx,
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(catalog.NewStatus(core.CatalogCacheStatus_CACHE_POPULATED, nil), tc.mockCacheServicePurError)

			mockCatalogClient := &catalogmocks.Client{}
			mockCatalogClient.On("Get",
				ctx,
				mock.Anything,
			).Return(tc.mockCatalogClientGetEntry, tc.mockCatalogClientGetError)

			fallbackClient, err := NewFallbackClient(mockCacheClient, mockCatalogClient)
			assert.NoError(t, err)

			entry, err := fallbackClient.Get(ctx, catalog.Key{})
			if tc.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, entry)
			}

			if tc.expectCatalogClientGetCall {
				mockCatalogClient.AssertCalled(t, "Get", mock.Anything, mock.Anything)
			} else {
				mockCatalogClient.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
			}

			if tc.expectCacheServicePutCall {
				mockCacheClient.AssertCalled(t, "Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			} else {
				mockCacheClient.AssertNotCalled(t, "Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}
