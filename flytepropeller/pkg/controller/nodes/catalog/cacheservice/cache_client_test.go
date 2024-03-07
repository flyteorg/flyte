package cacheservice

import (
	"context"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteidl/clients/go/cacheservice/mocks"
	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/cacheservice"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	mocks2 "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	pluginsIOMock "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	storageMocks "github.com/flyteorg/flyte/flytestdlib/storage/mocks"
)

var sampleIdentifier = core.Identifier{
	Project: "project_1",
	Domain:  "domain_1",
	Name:    "name_1",
	Version: "0",
	Org:     "org_1",
}

func generateCatalogKeys(ctx context.Context, t *testing.T) (catalog.Key, catalog.Key, string) {
	sampleInputs, err := coreutils.MakeLiteralMap(map[string]interface{}{"a": 1, "b": 2})
	assert.NoError(t, err)
	mockInputReader := &mocks2.InputReader{}
	mockInputReader.On("Get", mock.Anything).Return(sampleInputs, nil, nil)
	mockInputReaderErr := &mocks2.InputReader{}
	mockInputReaderErr.On("Get", mock.Anything).Return(sampleInputs, errors.New("test error"), nil)
	sampleInterface := core.TypedInterface{
		Inputs: &core.VariableMap{
			Variables: map[string]*core.Variable{
				"a": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
				"b": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
			},
		},
	}
	sampleCatalogKey := catalog.Key{
		Identifier:     sampleIdentifier,
		CacheVersion:   "1.0.0",
		TypedInterface: sampleInterface,
		InputReader:    mockInputReader,
	}
	sampleCatalogKeyErr := catalog.Key{
		Identifier:     sampleIdentifier,
		CacheVersion:   "1.0.0",
		TypedInterface: sampleInterface,
		InputReader:    mockInputReaderErr,
	}
	sampleCacheKey, err := GenerateCacheKey(ctx, sampleCatalogKey)
	assert.NoError(t, err)
	_, err = GenerateCacheKey(ctx, sampleCatalogKeyErr)
	assert.Error(t, err)

	return sampleCatalogKey, sampleCatalogKeyErr, sampleCacheKey
}

func TestCache_Get(t *testing.T) {
	ctx := context.Background()

	sampleCatalogKey, sampleCatalogKeyErr, sampleCacheKey := generateCatalogKeys(ctx, t)

	nonExpiredCreatedAt, err := ptypes.TimestampProto(time.Now().Add(time.Minute * 1))
	assert.NoError(t, err)
	expiredCreatedAt, err := ptypes.TimestampProto(time.Now().Add(time.Minute * -61))
	assert.NoError(t, err)
	sampleOutputLiteral, err := coreutils.MakeLiteralMap(map[string]interface{}{"c": 3})
	assert.NoError(t, err)
	outputLiteral := &cacheservice.CachedOutput_OutputLiterals{
		OutputLiterals: sampleOutputLiteral,
	}
	outputURI := &cacheservice.CachedOutput_OutputUri{
		OutputUri: "s3://some-bucket/some-key",
	}

	expiredResponse := &cacheservice.GetCacheResponse{
		Output: &cacheservice.CachedOutput{
			Output: &cacheservice.CachedOutput_OutputLiterals{},
			Metadata: &cacheservice.Metadata{
				CreatedAt: expiredCreatedAt,
			},
		},
	}
	responseLiteral := &cacheservice.GetCacheResponse{
		Output: &cacheservice.CachedOutput{
			Output: outputLiteral,
			Metadata: &cacheservice.Metadata{
				CreatedAt:        nonExpiredCreatedAt,
				SourceIdentifier: &sampleIdentifier,
			},
		},
	}
	responseURI := &cacheservice.GetCacheResponse{
		Output: &cacheservice.CachedOutput{
			Output: outputURI,
			Metadata: &cacheservice.Metadata{
				CreatedAt:        nonExpiredCreatedAt,
				SourceIdentifier: &sampleIdentifier,
			},
		},
	}
	malformedOutputResponse := &cacheservice.GetCacheResponse{
		Output: &cacheservice.CachedOutput{
			Metadata: &cacheservice.Metadata{
				CreatedAt:        nonExpiredCreatedAt,
				SourceIdentifier: &sampleIdentifier,
			},
		},
	}
	nilMetadataResponse := &cacheservice.GetCacheResponse{
		Output: &cacheservice.CachedOutput{
			Output: &cacheservice.CachedOutput_OutputUri{
				OutputUri: "s3://some-bucket/some-key",
			},
		},
	}
	malformedMetadataResponse := &cacheservice.GetCacheResponse{
		Output: &cacheservice.CachedOutput{
			Output: &cacheservice.CachedOutput_OutputUri{
				OutputUri: "s3://some-bucket/some-key",
			},
			Metadata: &cacheservice.Metadata{
				CreatedAt:        nonExpiredCreatedAt,
				SourceIdentifier: nil,
			},
		},
	}

	testCases := []struct {
		name                      string
		catalogKey                catalog.Key
		mockServiceReturnResponse *cacheservice.GetCacheResponse
		mockServiceReturnError    error
		readPbError               error
		expectClientError         bool
		expectedClientStatus      core.CatalogCacheStatus
		expectedClientErrCode     codes.Code
	}{
		{
			name:                      "generate cache key error",
			catalogKey:                sampleCatalogKeyErr,
			mockServiceReturnResponse: nil,
			mockServiceReturnError:    nil,
			expectClientError:         true,
			expectedClientStatus:      core.CatalogCacheStatus_CACHE_DISABLED,
			expectedClientErrCode:     codes.Unknown,
		},
		{
			name:                      "cache miss",
			catalogKey:                sampleCatalogKey,
			mockServiceReturnResponse: nil,
			mockServiceReturnError:    status.Error(codes.NotFound, "test not found"),
			expectClientError:         true,
			expectedClientStatus:      core.CatalogCacheStatus_CACHE_DISABLED,
			expectedClientErrCode:     codes.NotFound,
		},
		{
			name:                      "expired cache",
			catalogKey:                sampleCatalogKey,
			mockServiceReturnResponse: expiredResponse,
			mockServiceReturnError:    nil,
			expectClientError:         true,
			expectedClientStatus:      core.CatalogCacheStatus_CACHE_DISABLED,
			expectedClientErrCode:     codes.NotFound,
		},
		{
			name:                      "non-expired cache - literal",
			catalogKey:                sampleCatalogKey,
			mockServiceReturnResponse: responseLiteral,
			mockServiceReturnError:    nil,
			expectClientError:         false,
			expectedClientStatus:      core.CatalogCacheStatus_CACHE_HIT,
		},
		{
			name:                      "non-expired cache - uri",
			catalogKey:                sampleCatalogKey,
			mockServiceReturnResponse: responseURI,
			mockServiceReturnError:    nil,
			expectClientError:         false,
			expectedClientStatus:      core.CatalogCacheStatus_CACHE_HIT,
		},
		{
			name:                      "non-expired cache - uri - read proto error",
			catalogKey:                sampleCatalogKey,
			mockServiceReturnResponse: responseURI,
			mockServiceReturnError:    nil,
			readPbError:               errors.New("test error"),
			expectClientError:         true,
			expectedClientStatus:      core.CatalogCacheStatus_CACHE_DISABLED,
			expectedClientErrCode:     codes.Unknown,
		},
		{
			name:                      "malformed output response",
			catalogKey:                sampleCatalogKey,
			mockServiceReturnResponse: malformedOutputResponse,
			mockServiceReturnError:    nil,
			expectClientError:         true,
			expectedClientStatus:      core.CatalogCacheStatus_CACHE_DISABLED,
			expectedClientErrCode:     codes.Internal,
		},
		{
			name:                      "nil metadata response",
			catalogKey:                sampleCatalogKey,
			mockServiceReturnResponse: nilMetadataResponse,
			mockServiceReturnError:    nil,
			expectClientError:         true,
			expectedClientStatus:      core.CatalogCacheStatus_CACHE_DISABLED,
			expectedClientErrCode:     codes.Internal,
		},
		{
			name:                      "malformed metadata response",
			catalogKey:                sampleCatalogKey,
			mockServiceReturnResponse: malformedMetadataResponse,
			mockServiceReturnError:    nil,
			expectClientError:         true,
			expectedClientStatus:      core.CatalogCacheStatus_CACHE_DISABLED,
			expectedClientErrCode:     codes.Unknown,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPBStore := &storageMocks.ComposedProtobufStore{}
			mockPBStore.On("ReadProtobuf", mock.Anything, mock.Anything, mock.Anything).Return(tc.readPbError)
			store := &storage.DataStore{
				ComposedProtobufStore: mockPBStore,
				ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
			}
			mockServiceClient := &mocks.CacheServiceClient{}
			mockServiceClient.On("Get",
				ctx,
				mock.MatchedBy(func(o *cacheservice.GetCacheRequest) bool {
					assert.EqualValues(t, sampleCacheKey, o.Key)
					return true
				}),
			).Return(tc.mockServiceReturnResponse, tc.mockServiceReturnError)

			cacheClient := &CacheClient{
				client:      mockServiceClient,
				maxCacheAge: time.Hour,
				store:       store,
			}

			resp, err := cacheClient.Get(ctx, tc.catalogKey)
			if tc.expectClientError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedClientErrCode, status.Code(errors.Cause(err)))
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
			assert.Equal(t, tc.expectedClientStatus, resp.GetStatus().GetCacheStatus())
		})
	}
}

type MockStoreMetadata struct {
	exists bool
	size   int64
	etag   string
}

func (m MockStoreMetadata) Size() int64 {
	return m.size
}

func (m MockStoreMetadata) Exists() bool {
	return m.exists
}

func (m MockStoreMetadata) Etag() string {
	return m.etag
}

func TestCache_Put(t *testing.T) {
	ctx := context.Background()

	sampleCatalogKey, sampleCatalogKeyErr, sampleCacheKey := generateCatalogKeys(ctx, t)

	sampleOutputs, err := coreutils.MakeLiteralMap(map[string]interface{}{"c": 3})
	assert.NoError(t, err)

	mockOutputReader := &mocks2.OutputReader{}
	mockOutputReader.On("Read", mock.Anything).Return(sampleOutputs, nil, nil)
	mockOutputReader.On("GetOutputPath", mock.Anything).Return(storage.DataReference("s3://some-bucket/some-key"), nil)

	mockOutputReaderExecErr := &mocks2.OutputReader{}
	mockOutputReaderExecErr.On("Read", mock.Anything).Return(nil, &io.ExecutionError{}, nil)

	mockOutputReaderErr := &mocks2.OutputReader{}
	mockOutputReaderErr.On("Read", mock.Anything).Return(nil, nil, errors.New("test error"))

	opath := &pluginsIOMock.OutputFilePaths{}
	opath.On("GetOutputPath").Return(storage.DataReference("s3://some-bucket/some-key"))

	composedProtobufStoreExists := &storageMocks.ComposedProtobufStore{}
	composedProtobufStoreExists.On("Head", mock.Anything, mock.Anything).Return(MockStoreMetadata{exists: true}, nil)
	mockRemoteFileOutputReader := ioutils.NewRemoteFileOutputReader(ctx, composedProtobufStoreExists, opath, 0)

	composedProtobufStoreNotExists := &storageMocks.ComposedProtobufStore{}
	composedProtobufStoreNotExists.On("Head", mock.Anything, mock.Anything).Return(MockStoreMetadata{exists: false}, nil)
	mockRemoteFileOutputReaderErr := ioutils.NewRemoteFileOutputReader(ctx, composedProtobufStoreNotExists, opath, 0)

	mockMetadata := catalog.Metadata{
		TaskExecutionIdentifier: &core.TaskExecutionIdentifier{},
	}

	testCases := []struct {
		name                      string
		catalogKey                catalog.Key
		mockOutputReader          io.OutputReader
		mockServiceReturnResponse *cacheservice.PutCacheResponse
		mockServiceReturnError    error
		expectClientError         bool
		expectedClientStatus      core.CatalogCacheStatus
		expectedClientErrCode     codes.Code
		inlineCache               bool
	}{
		{
			name:                      "successful put, literal",
			catalogKey:                sampleCatalogKey,
			mockOutputReader:          mockOutputReader,
			mockServiceReturnResponse: &cacheservice.PutCacheResponse{},
			mockServiceReturnError:    nil,
			expectedClientStatus:      core.CatalogCacheStatus_CACHE_POPULATED,
			inlineCache:               true,
		},
		{
			name:                      "output reader exec error",
			catalogKey:                sampleCatalogKey,
			mockOutputReader:          mockOutputReaderExecErr,
			mockServiceReturnResponse: &cacheservice.PutCacheResponse{},
			mockServiceReturnError:    nil,
			expectedClientStatus:      core.CatalogCacheStatus_CACHE_DISABLED,
			expectClientError:         true,
			inlineCache:               true,
		},
		{
			name:                      "output reader error",
			catalogKey:                sampleCatalogKey,
			mockOutputReader:          mockOutputReaderErr,
			mockServiceReturnResponse: &cacheservice.PutCacheResponse{},
			mockServiceReturnError:    nil,
			expectedClientStatus:      core.CatalogCacheStatus_CACHE_DISABLED,
			expectClientError:         true,
			inlineCache:               true,
		},
		{
			name:                      "cache service error",
			catalogKey:                sampleCatalogKey,
			mockOutputReader:          mockOutputReader,
			mockServiceReturnResponse: nil,
			mockServiceReturnError:    status.Error(codes.NotFound, "test error"),
			expectedClientStatus:      core.CatalogCacheStatus_CACHE_DISABLED,
			expectClientError:         true,
			inlineCache:               true,
		},
		{
			name:                      "generate cache key error",
			catalogKey:                sampleCatalogKeyErr,
			mockOutputReader:          mockOutputReader,
			mockServiceReturnResponse: nil,
			mockServiceReturnError:    nil,
			expectedClientStatus:      core.CatalogCacheStatus_CACHE_DISABLED,
			expectClientError:         true,
			inlineCache:               true,
		},
		{
			name:                      "successful put, uri",
			catalogKey:                sampleCatalogKey,
			mockOutputReader:          mockRemoteFileOutputReader,
			mockServiceReturnResponse: &cacheservice.PutCacheResponse{},
			mockServiceReturnError:    nil,
			expectedClientStatus:      core.CatalogCacheStatus_CACHE_POPULATED,
			inlineCache:               false,
		},
		{
			name:                 "output reader error - uri path does not exist",
			catalogKey:           sampleCatalogKey,
			mockOutputReader:     mockRemoteFileOutputReaderErr,
			expectedClientStatus: core.CatalogCacheStatus_CACHE_DISABLED,
			expectClientError:    true,
			inlineCache:          false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockServiceClient := &mocks.CacheServiceClient{}
			mockServiceClient.On("Put",
				ctx,
				mock.MatchedBy(func(o *cacheservice.PutCacheRequest) bool {
					assert.EqualValues(t, sampleCacheKey, o.Key)
					assert.EqualValues(t, false, o.Overwrite)
					if tc.inlineCache {
						_, ok := o.Output.Output.(*cacheservice.CachedOutput_OutputLiterals)
						assert.True(t, ok, "Expected output to be of type *cacheservice.CachedOutput_OutputLiterals")
					} else {
						_, ok := o.Output.Output.(*cacheservice.CachedOutput_OutputUri)
						assert.True(t, ok, "Expected output to be of type *cacheservice.CachedOutput_OutputUri")
					}
					return true
				}),
			).Return(tc.mockServiceReturnResponse, tc.mockServiceReturnError)

			cacheClient := &CacheClient{
				client:      mockServiceClient,
				maxCacheAge: time.Hour,
				inlineCache: tc.inlineCache,
			}
			resp, err := cacheClient.Put(ctx, tc.catalogKey, tc.mockOutputReader, mockMetadata)

			if tc.expectClientError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
			assert.Equal(t, tc.expectedClientStatus, resp.GetCacheStatus())
		})
	}

	t.Run("Test Update overwrites", func(t *testing.T) {
		mockClient := &mocks.CacheServiceClient{}
		cacheClient := &CacheClient{
			client:      mockClient,
			maxCacheAge: time.Hour,
			inlineCache: true,
		}
		mockClient.On("Put",
			ctx,
			mock.MatchedBy(func(o *cacheservice.PutCacheRequest) bool {
				assert.EqualValues(t, sampleCacheKey, o.Key)
				assert.EqualValues(t, true, o.Overwrite)
				return true
			}),
		).Return(&cacheservice.PutCacheResponse{}, nil)
		_, _ = cacheClient.Update(ctx, sampleCatalogKey, mockOutputReader, mockMetadata)
	})
}

func TestCache_ReleaseReservation(t *testing.T) {
	ctx := context.Background()

	sampleCatalogKey, sampleCatalogKeyErr, sampleCacheKey := generateCatalogKeys(ctx, t)
	owner := "owner"

	testCases := []struct {
		name                      string
		catalogKey                catalog.Key
		mockServiceReturnResponse *cacheservice.ReleaseReservationResponse
		mockServiceReturnError    error
		expectClientError         bool
		expectedClientErrCode     codes.Code
	}{
		{
			name:                      "generate cache key error",
			catalogKey:                sampleCatalogKeyErr,
			mockServiceReturnResponse: nil,
			mockServiceReturnError:    nil,
			expectClientError:         true,
			expectedClientErrCode:     codes.Unknown,
		},
		{
			name:                      "release reservation error",
			catalogKey:                sampleCatalogKey,
			mockServiceReturnResponse: nil,
			mockServiceReturnError:    status.Error(codes.Internal, "release reservation error"),
			expectClientError:         true,
			expectedClientErrCode:     codes.Internal,
		},
		{
			name:                      "successful release",
			catalogKey:                sampleCatalogKey,
			mockServiceReturnResponse: nil,
			mockServiceReturnError:    nil,
			expectClientError:         false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockServiceClient := &mocks.CacheServiceClient{}
			mockServiceClient.On("ReleaseReservation",
				ctx,
				mock.MatchedBy(func(r *cacheservice.ReleaseReservationRequest) bool {
					assert.Equal(t, sampleCacheKey, r.Key)
					assert.Equal(t, owner, r.OwnerId)
					return true
				}),
			).Return(tc.mockServiceReturnResponse, tc.mockServiceReturnError)

			cacheClient := &CacheClient{
				client:      mockServiceClient,
				maxCacheAge: time.Hour,
			}
			err := cacheClient.ReleaseReservation(ctx, tc.catalogKey, owner)

			if tc.expectClientError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedClientErrCode, status.Code(errors.Cause(err)))
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestCache_GetOrExtendReservation(t *testing.T) {
	ctx := context.Background()

	sampleCatalogKey, sampleCatalogKeyErr, sampleCacheKey := generateCatalogKeys(ctx, t)
	owner := "owner"

	testCases := []struct {
		name                      string
		catalogKey                catalog.Key
		mockServiceReturnResponse *cacheservice.GetOrExtendReservationResponse
		mockServiceReturnError    error
		expectClientError         bool
		expectedClientErrCode     codes.Code
	}{
		{
			name:                      "generate cache key error",
			catalogKey:                sampleCatalogKeyErr,
			mockServiceReturnResponse: nil,
			mockServiceReturnError:    nil,
			expectClientError:         true,
			expectedClientErrCode:     codes.Unknown,
		},
		{
			name:                      "get or extend reservation error",
			catalogKey:                sampleCatalogKey,
			mockServiceReturnResponse: nil,
			mockServiceReturnError:    status.Error(codes.Internal, "get or extend reservation error"),
			expectClientError:         true,
			expectedClientErrCode:     codes.Internal,
		},
		{
			name:       "successful get or extend reservation",
			catalogKey: sampleCatalogKey,
			mockServiceReturnResponse: &cacheservice.GetOrExtendReservationResponse{
				Reservation: &cacheservice.Reservation{
					OwnerId:           owner,
					ExpiresAt:         ptypes.TimestampNow(),
					HeartbeatInterval: ptypes.DurationProto(time.Minute * 1),
				},
			},
			mockServiceReturnError: nil,
			expectClientError:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockServiceClient := &mocks.CacheServiceClient{}
			mockServiceClient.On("GetOrExtendReservation",
				ctx,
				mock.MatchedBy(func(r *cacheservice.GetOrExtendReservationRequest) bool {
					assert.Equal(t, sampleCacheKey, r.Key)
					assert.Equal(t, owner, r.OwnerId)
					return true
				}),
			).Return(tc.mockServiceReturnResponse, tc.mockServiceReturnError)

			cacheClient := &CacheClient{
				client:      mockServiceClient,
				maxCacheAge: time.Hour,
			}
			reservation, err := cacheClient.GetOrExtendReservation(ctx, tc.catalogKey, owner, time.Minute*1)

			if tc.expectClientError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedClientErrCode, status.Code(errors.Cause(err)))
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reservation)
			}
		})
	}
}
