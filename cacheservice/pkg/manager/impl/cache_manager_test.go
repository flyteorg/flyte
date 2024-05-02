package impl

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/flyteorg/flyte/cacheservice/pkg/errors"
	"github.com/flyteorg/flyte/cacheservice/pkg/manager/interfaces"
	"github.com/flyteorg/flyte/cacheservice/pkg/manager/mocks"
	repoMocks "github.com/flyteorg/flyte/cacheservice/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/cacheservice/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/cacheservice"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

var sampleKey = "key"

func init() {
	labeled.SetMetricKeys(contextutils.AppNameKey)
}

func TestCacheManager_Get(t *testing.T) {
	ctx := context.Background()

	sampleOutput := &models.CachedOutput{}
	malformedOutput := &models.CachedOutput{
		SerializedMetadata: []byte("malformed"),
	}

	getError := errors.NewCacheServiceError(codes.Internal, "get error")
	notFoundError := errors.NewNotFoundError("output", sampleKey)

	testCases := []struct {
		name                    string
		mockOutputStore         interfaces.CacheOutputBlobStore
		mockRequest             *cacheservice.GetCacheRequest
		mockDataStoreReturn     *models.CachedOutput
		mockDataStoreError      error
		expectError             bool
		expectedErrorStatusCode codes.Code
	}{
		{
			name:                    "failed validation",
			mockRequest:             &cacheservice.GetCacheRequest{},
			expectError:             true,
			expectedErrorStatusCode: codes.InvalidArgument,
		},
		{
			name:                    "get error",
			mockRequest:             &cacheservice.GetCacheRequest{Key: sampleKey},
			mockDataStoreReturn:     nil,
			mockDataStoreError:      getError,
			expectError:             true,
			expectedErrorStatusCode: codes.Internal,
		},
		{
			name:                    "not found error",
			mockRequest:             &cacheservice.GetCacheRequest{Key: sampleKey},
			mockDataStoreReturn:     nil,
			mockDataStoreError:      notFoundError,
			expectError:             true,
			expectedErrorStatusCode: codes.NotFound,
		},
		{
			name:                    "malformed output",
			mockRequest:             &cacheservice.GetCacheRequest{Key: sampleKey},
			mockDataStoreReturn:     malformedOutput,
			mockDataStoreError:      nil,
			expectError:             true,
			expectedErrorStatusCode: codes.Internal,
		},
		{
			name:                "success",
			mockRequest:         &cacheservice.GetCacheRequest{Key: sampleKey},
			mockDataStoreReturn: sampleOutput,
			mockDataStoreError:  nil,
			expectError:         false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDataStore := &repoMocks.CachedOutputRepo{}
			mockDataStore.OnGetMatch(
				ctx,
				mock.MatchedBy(func(o string) bool {
					assert.EqualValues(t, sampleKey, o)
					return true
				})).Return(tc.mockDataStoreReturn, tc.mockDataStoreError)

			m := NewCacheManager(&mocks.CacheOutputBlobStore{}, mockDataStore, &repoMocks.ReservationRepo{}, 1024, mockScope.NewTestScope(), time.Duration(1), time.Duration(1))

			response, err := m.Get(ctx, tc.mockRequest)

			if tc.expectError {
				assert.Error(t, err)
				assert.EqualValues(t, tc.expectedErrorStatusCode, status.Code(err))
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
			}
		})
	}
}

func TestCacheManager_Put(t *testing.T) {
	ctx := context.Background()

	notFoundError := errors.NewNotFoundError("output", sampleKey)
	sampleOutputURI := "outputUri"
	sampleOutputLiteral, err := coreutils.MakeLiteralMap(map[string]interface{}{"c": 3})
	assert.NoError(t, err)

	validSourceIdentifier := &core.Identifier{}
	validMetadata := &cacheservice.Metadata{SourceIdentifier: validSourceIdentifier}

	matchKeyFunc := func(key string) bool {
		return assert.EqualValues(t, sampleKey, key)
	}

	testCases := []struct {
		name                        string
		mockRequest                 *cacheservice.PutCacheRequest
		dataGetReturn               *models.CachedOutput
		dataGetError                error
		dataPutMatchOutputFunc      func(*models.CachedOutput) bool
		dataPutError                error
		outputCreateMatchOutputFunc func(*core.LiteralMap) bool
		outputCreateReturn          string
		outputCreateError           error
		outputDeleteMatchOutputFunc func(string) bool
		outputDeleteError           error
		maxSizeBytes                int64
		expectError                 bool
		expectOutputDelete          bool
		expectedErrorStatusCode     codes.Code
	}{
		{
			name:                    "failed validation",
			mockRequest:             &cacheservice.PutCacheRequest{},
			expectError:             true,
			expectedErrorStatusCode: codes.InvalidArgument,
		},
		{
			name: "output URI, not found, success",
			mockRequest: &cacheservice.PutCacheRequest{
				Key: sampleKey,
				Output: &cacheservice.CachedOutput{
					Output: &cacheservice.CachedOutput_OutputUri{
						OutputUri: sampleOutputURI,
					},
					Metadata: validMetadata,
				},
				Overwrite: &cacheservice.OverwriteOutput{
					Overwrite:  false,
					DeleteBlob: false,
				},
			},
			dataGetReturn: nil,
			dataGetError:  notFoundError,
			dataPutMatchOutputFunc: func(cachedOutput *models.CachedOutput) bool {
				return assert.EqualValues(t, sampleOutputURI, cachedOutput.OutputURI) && assert.Nil(t, cachedOutput.OutputLiteral)
			},
			dataPutError: nil,
			expectError:  false,
		},
		{
			name: "output URI, error",
			mockRequest: &cacheservice.PutCacheRequest{
				Key: sampleKey,
				Output: &cacheservice.CachedOutput{
					Output: &cacheservice.CachedOutput_OutputUri{
						OutputUri: sampleOutputURI,
					},
					Metadata: validMetadata,
				},
				Overwrite: &cacheservice.OverwriteOutput{
					Overwrite:  false,
					DeleteBlob: false,
				},
			},
			dataGetReturn: nil,
			dataGetError:  notFoundError,
			dataPutMatchOutputFunc: func(cachedOutput *models.CachedOutput) bool {
				return assert.EqualValues(t, sampleOutputURI, cachedOutput.OutputURI) && assert.Nil(t, cachedOutput.OutputLiteral)
			},
			dataPutError:            errors.NewCacheServiceErrorf(codes.Internal, "data put error"),
			expectError:             true,
			expectedErrorStatusCode: codes.Internal,
		},
		{
			name: "output literals, inline, success",
			mockRequest: &cacheservice.PutCacheRequest{
				Key: sampleKey,
				Output: &cacheservice.CachedOutput{
					Output: &cacheservice.CachedOutput_OutputLiterals{
						OutputLiterals: sampleOutputLiteral,
					},
					Metadata: validMetadata,
				},
			},
			dataGetReturn: nil,
			dataGetError:  notFoundError,
			dataPutMatchOutputFunc: func(cachedOutput *models.CachedOutput) bool {
				return assert.NotNil(t, cachedOutput.OutputLiteral) && assert.EqualValues(t, "", cachedOutput.OutputURI)
			},
			dataPutError: nil,
			maxSizeBytes: 1000000,
			expectError:  false,
		},
		{
			name: "output literals, inline, error",
			mockRequest: &cacheservice.PutCacheRequest{
				Key: sampleKey,
				Output: &cacheservice.CachedOutput{
					Output: &cacheservice.CachedOutput_OutputLiterals{
						OutputLiterals: sampleOutputLiteral,
					},
					Metadata: validMetadata,
				},
			},
			dataGetReturn: nil,
			dataGetError:  notFoundError,
			dataPutMatchOutputFunc: func(cachedOutput *models.CachedOutput) bool {
				return assert.NotNil(t, cachedOutput.OutputLiteral) && assert.EqualValues(t, "", cachedOutput.OutputURI)
			},
			dataPutError:            errors.NewCacheServiceErrorf(codes.Internal, "data put error"),
			maxSizeBytes:            1000000,
			expectError:             true,
			expectedErrorStatusCode: codes.Internal,
		},
		{
			name: "output literals, offload, success ",
			mockRequest: &cacheservice.PutCacheRequest{
				Key: sampleKey,
				Output: &cacheservice.CachedOutput{
					Output: &cacheservice.CachedOutput_OutputLiterals{
						OutputLiterals: sampleOutputLiteral,
					},
					Metadata: validMetadata,
				},
			},
			dataGetReturn: nil,
			dataGetError:  notFoundError,
			dataPutMatchOutputFunc: func(cachedOutput *models.CachedOutput) bool {
				return assert.EqualValues(t, sampleOutputURI, cachedOutput.OutputURI) && assert.Nil(t, cachedOutput.OutputLiteral)
			},
			dataPutError: nil,
			outputCreateMatchOutputFunc: func(outputLiteral *core.LiteralMap) bool {
				return assert.EqualValues(t, sampleOutputLiteral, outputLiteral)
			},
			outputCreateReturn: sampleOutputURI,
			outputCreateError:  nil,
			maxSizeBytes:       1,
			expectError:        false,
		},
		{
			name: "output literals, offload, error",
			mockRequest: &cacheservice.PutCacheRequest{
				Key: sampleKey,
				Output: &cacheservice.CachedOutput{
					Output: &cacheservice.CachedOutput_OutputLiterals{
						OutputLiterals: sampleOutputLiteral,
					},
					Metadata: validMetadata,
				},
			},
			dataGetReturn: nil,
			dataGetError:  notFoundError,
			outputCreateMatchOutputFunc: func(outputLiteral *core.LiteralMap) bool {
				return assert.EqualValues(t, sampleOutputLiteral, outputLiteral)
			},
			outputCreateReturn:      sampleOutputURI,
			outputCreateError:       errors.NewCacheServiceError(codes.Internal, "output create error"),
			maxSizeBytes:            1,
			expectError:             true,
			expectedErrorStatusCode: codes.Internal,
		},
		{
			name: "output exists, don't overwrite, error",
			mockRequest: &cacheservice.PutCacheRequest{
				Key: sampleKey,
				Output: &cacheservice.CachedOutput{
					Output: &cacheservice.CachedOutput_OutputUri{
						OutputUri: sampleOutputURI,
					},
					Metadata: validMetadata,
				},
				Overwrite: &cacheservice.OverwriteOutput{
					Overwrite:  false,
					DeleteBlob: false,
				},
			},
			dataGetReturn:           &models.CachedOutput{},
			dataGetError:            nil,
			expectError:             true,
			expectedErrorStatusCode: codes.AlreadyExists,
		},
		{
			name: "output exists, overwrite unset, error",
			mockRequest: &cacheservice.PutCacheRequest{
				Key: sampleKey,
				Output: &cacheservice.CachedOutput{
					Output: &cacheservice.CachedOutput_OutputUri{
						OutputUri: sampleOutputURI,
					},
					Metadata: validMetadata,
				},
			},
			dataGetReturn:           &models.CachedOutput{},
			dataGetError:            nil,
			expectError:             true,
			expectedErrorStatusCode: codes.AlreadyExists,
		},
		{
			name: "output exists, overwrite + delete, success",
			mockRequest: &cacheservice.PutCacheRequest{
				Key: sampleKey,
				Output: &cacheservice.CachedOutput{
					Output: &cacheservice.CachedOutput_OutputUri{
						OutputUri: sampleOutputURI,
					},
					Metadata: validMetadata,
				},
				Overwrite: &cacheservice.OverwriteOutput{
					Overwrite:  true,
					DeleteBlob: true,
				},
			},
			dataGetReturn: &models.CachedOutput{
				OutputURI: sampleOutputURI,
			},
			dataGetError: nil,
			dataPutMatchOutputFunc: func(cachedOutput *models.CachedOutput) bool {
				return assert.EqualValues(t, sampleOutputURI, cachedOutput.OutputURI) && assert.Nil(t, cachedOutput.OutputLiteral)
			},
			outputDeleteMatchOutputFunc: func(outputUri string) bool {
				return assert.EqualValues(t, sampleOutputURI, outputUri)
			},
			outputDeleteError:  nil,
			outputCreateError:  nil,
			dataPutError:       nil,
			expectError:        false,
			expectOutputDelete: true,
		},
		{
			name: "output exists, overwrite + no delete, success",
			mockRequest: &cacheservice.PutCacheRequest{
				Key: sampleKey,
				Output: &cacheservice.CachedOutput{
					Output: &cacheservice.CachedOutput_OutputUri{
						OutputUri: sampleOutputURI,
					},
					Metadata: validMetadata,
				},
				Overwrite: &cacheservice.OverwriteOutput{
					Overwrite:  true,
					DeleteBlob: false,
				},
			},
			dataGetReturn: &models.CachedOutput{
				OutputURI: sampleOutputURI,
			},
			dataGetError: nil,
			dataPutMatchOutputFunc: func(cachedOutput *models.CachedOutput) bool {
				return assert.EqualValues(t, sampleOutputURI, cachedOutput.OutputURI) && assert.Nil(t, cachedOutput.OutputLiteral)
			},
			outputDeleteError:  nil,
			outputCreateError:  nil,
			dataPutError:       nil,
			expectError:        false,
			expectOutputDelete: false,
		},
		{
			name: "output exists, delete fails, error",
			mockRequest: &cacheservice.PutCacheRequest{
				Key: sampleKey,
				Output: &cacheservice.CachedOutput{
					Output: &cacheservice.CachedOutput_OutputUri{
						OutputUri: sampleOutputURI,
					},
					Metadata: validMetadata,
				},
				Overwrite: &cacheservice.OverwriteOutput{
					Overwrite:  true,
					DeleteBlob: true,
				},
			},
			dataGetReturn: &models.CachedOutput{
				OutputURI: sampleOutputURI,
			},
			dataGetError: nil,
			dataPutMatchOutputFunc: func(cachedOutput *models.CachedOutput) bool {
				return assert.EqualValues(t, sampleOutputURI, cachedOutput.OutputURI) && assert.Nil(t, cachedOutput.OutputLiteral)
			},
			outputDeleteMatchOutputFunc: func(outputUri string) bool {
				return assert.EqualValues(t, sampleOutputURI, outputUri)
			},
			outputDeleteError:       errors.NewCacheServiceErrorf(codes.Internal, "output delete error"),
			expectError:             true,
			expectOutputDelete:      true,
			expectedErrorStatusCode: codes.Internal,
		},
		{
			name: "get error, error",
			mockRequest: &cacheservice.PutCacheRequest{
				Key: sampleKey,
				Output: &cacheservice.CachedOutput{
					Output: &cacheservice.CachedOutput_OutputUri{
						OutputUri: sampleOutputURI,
					},
					Metadata: validMetadata,
				},
				Overwrite: &cacheservice.OverwriteOutput{
					Overwrite:  false,
					DeleteBlob: false,
				},
			},
			dataGetReturn:           nil,
			dataGetError:            errors.NewCacheServiceError(codes.Internal, "get output error"),
			expectError:             true,
			expectedErrorStatusCode: codes.Internal,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDataStore := &repoMocks.CachedOutputRepo{}
			mockDataStore.OnGetMatch(
				ctx,
				mock.MatchedBy(matchKeyFunc),
			).Return(tc.dataGetReturn, tc.dataGetError)
			mockDataStore.OnPutMatch(
				ctx,
				mock.MatchedBy(matchKeyFunc),
				mock.MatchedBy(tc.dataPutMatchOutputFunc),
			).Return(tc.dataPutError)

			mockOutputStore := &mocks.CacheOutputBlobStore{}
			mockOutputStore.OnCreateMatch(
				ctx,
				mock.MatchedBy(matchKeyFunc),
				mock.MatchedBy(tc.outputCreateMatchOutputFunc),
			).Return(tc.outputCreateReturn, tc.outputCreateError)
			mockOutputStore.OnDeleteMatch(
				ctx,
				mock.MatchedBy(tc.outputDeleteMatchOutputFunc),
			).Return(tc.outputDeleteError)

			m := NewCacheManager(mockOutputStore, mockDataStore, &repoMocks.ReservationRepo{}, tc.maxSizeBytes, mockScope.NewTestScope(), time.Duration(1), time.Duration(1))

			_, err := m.Put(ctx, tc.mockRequest)

			if tc.expectError {
				assert.Error(t, err)
				assert.EqualValues(t, tc.expectedErrorStatusCode, status.Code(err))
			} else {
				assert.NoError(t, err)
			}

			if tc.expectOutputDelete {
				mockOutputStore.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestCacheManager_GetOrExtendReservation(t *testing.T) {
	ctx := context.Background()

	heartbeatGracePeriodMultiplier := time.Duration(3)
	maxHeartBeatInterval := durationpb.New(2 * time.Second).AsDuration()

	sampleOwner := "owner"
	sampleOwner1 := "owner1"
	sampleHeartBeatInterval := durationpb.New(1 * time.Second)

	now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name                    string
		mockRequest             *cacheservice.GetOrExtendReservationRequest
		expectError             bool
		expectedErrorStatusCode codes.Code
		deleteError             error
		getReturn               *models.CacheReservation
		getError                error
		getReturn1              *models.CacheReservation
		getError1               error
		createError             error
		updateError             error
		expectCreateCalled      bool
		expectUpdateCalled      bool
		expectedExpiresAtCall   time.Time
	}{
		{
			name:                    "failed validation",
			mockRequest:             &cacheservice.GetOrExtendReservationRequest{},
			expectError:             true,
			expectedErrorStatusCode: codes.InvalidArgument,
		},
		{
			name: "no existing reservation, create",
			mockRequest: &cacheservice.GetOrExtendReservationRequest{
				Key:               sampleKey,
				OwnerId:           sampleOwner,
				HeartbeatInterval: sampleHeartBeatInterval,
			},
			getReturn:             nil,
			getError:              errors.NewNotFoundError("reservation", sampleKey),
			createError:           nil,
			expectError:           false,
			expectCreateCalled:    true,
			expectedExpiresAtCall: now.Add(sampleHeartBeatInterval.AsDuration() * heartbeatGracePeriodMultiplier),
		},
		{
			name: "exceeds max heartbeat interval",
			mockRequest: &cacheservice.GetOrExtendReservationRequest{
				Key:               sampleKey,
				OwnerId:           sampleOwner,
				HeartbeatInterval: durationpb.New(maxHeartBeatInterval * 10),
			},
			getReturn:             nil,
			getError:              errors.NewNotFoundError("reservation", sampleKey),
			createError:           nil,
			expectError:           false,
			expectCreateCalled:    true,
			expectedExpiresAtCall: now.Add(maxHeartBeatInterval * heartbeatGracePeriodMultiplier),
		},
		{
			name: "create fails - error",
			mockRequest: &cacheservice.GetOrExtendReservationRequest{
				Key:               sampleKey,
				OwnerId:           sampleOwner,
				HeartbeatInterval: sampleHeartBeatInterval,
			},
			getReturn:               nil,
			getError:                errors.NewNotFoundError("reservation", sampleKey),
			createError:             errors.NewCacheServiceErrorf(codes.Internal, "create error"),
			expectError:             true,
			expectedErrorStatusCode: codes.Internal,
			expectCreateCalled:      true,
			expectedExpiresAtCall:   now.Add(sampleHeartBeatInterval.AsDuration() * heartbeatGracePeriodMultiplier),
		},
		{
			name: "create fails - already exists",
			mockRequest: &cacheservice.GetOrExtendReservationRequest{
				Key:               sampleKey,
				OwnerId:           sampleOwner,
				HeartbeatInterval: sampleHeartBeatInterval,
			},
			getReturn: nil,
			getError:  errors.NewNotFoundError("reservation", sampleKey),
			getReturn1: &models.CacheReservation{
				Key:       sampleKey,
				OwnerID:   sampleOwner1,
				ExpiresAt: now.Add(time.Hour),
			},
			getError1:             nil,
			createError:           errors.NewCacheServiceErrorf(codes.AlreadyExists, "create error"),
			expectError:           false,
			expectCreateCalled:    true,
			expectedExpiresAtCall: now.Add(sampleHeartBeatInterval.AsDuration() * heartbeatGracePeriodMultiplier),
		},
		{
			name: "expired reservation, update",
			mockRequest: &cacheservice.GetOrExtendReservationRequest{
				Key:               sampleKey,
				OwnerId:           sampleOwner,
				HeartbeatInterval: sampleHeartBeatInterval,
			},
			getReturn: &models.CacheReservation{
				Key:       sampleKey,
				OwnerID:   sampleOwner1,
				ExpiresAt: now.Add(-time.Hour),
			},
			getError:              nil,
			updateError:           nil,
			expectError:           false,
			expectUpdateCalled:    true,
			expectedExpiresAtCall: now.Add(sampleHeartBeatInterval.AsDuration() * heartbeatGracePeriodMultiplier),
		},
		{
			name: "active reservation with same owner, update",
			mockRequest: &cacheservice.GetOrExtendReservationRequest{
				Key:               sampleKey,
				OwnerId:           sampleOwner,
				HeartbeatInterval: sampleHeartBeatInterval,
			},
			getReturn: &models.CacheReservation{
				Key:       sampleKey,
				OwnerID:   sampleOwner,
				ExpiresAt: now.Add(time.Hour),
			},
			getError:              nil,
			updateError:           nil,
			expectError:           false,
			expectUpdateCalled:    true,
			expectedExpiresAtCall: now.Add(sampleHeartBeatInterval.AsDuration() * heartbeatGracePeriodMultiplier),
		},
		{
			name: "active reservation, different owner",
			mockRequest: &cacheservice.GetOrExtendReservationRequest{
				Key:               sampleKey,
				OwnerId:           sampleOwner,
				HeartbeatInterval: sampleHeartBeatInterval,
			},
			getReturn: &models.CacheReservation{
				Key:       sampleKey,
				OwnerID:   sampleOwner1,
				ExpiresAt: now.Add(time.Hour),
			},
			getError:              nil,
			updateError:           nil,
			expectError:           false,
			expectUpdateCalled:    false,
			expectCreateCalled:    false,
			expectedExpiresAtCall: now.Add(sampleHeartBeatInterval.AsDuration() * heartbeatGracePeriodMultiplier),
		},
	}

	for _, tc := range testCases {
		requestKey := fmt.Sprintf("%s:%s", "reservation", tc.mockRequest.Key)
		t.Run(tc.name, func(t *testing.T) {
			mockReservationStoreClient := &repoMocks.ReservationRepo{}
			mockReservationStoreClient.OnGetMatch(
				ctx,
				mock.MatchedBy(func(o string) bool {
					assert.Equal(t, requestKey, o)
					return true
				})).Return(tc.getReturn, tc.getError).Once()
			mockReservationStoreClient.OnGetMatch(
				ctx,
				mock.MatchedBy(func(o string) bool {
					assert.Equal(t, requestKey, o)
					return true
				})).Return(tc.getReturn1, tc.getError1)
			mockReservationStoreClient.OnCreateMatch(
				ctx,
				mock.MatchedBy(func(reservation *models.CacheReservation) bool {
					assert.Equal(t, requestKey, reservation.Key)
					assert.Equal(t, tc.expectedExpiresAtCall, reservation.ExpiresAt)
					return true
				}),
				mock.Anything,
			).Return(tc.createError)
			mockReservationStoreClient.OnUpdateMatch(
				ctx,
				mock.MatchedBy(func(reservation *models.CacheReservation) bool {
					assert.Equal(t, requestKey, reservation.Key)
					assert.Equal(t, tc.expectedExpiresAtCall, reservation.ExpiresAt)
					return true
				}),
				mock.Anything,
			).Return(tc.updateError)

			m := NewCacheManager(&mocks.CacheOutputBlobStore{}, &repoMocks.CachedOutputRepo{}, mockReservationStoreClient, 0, mockScope.NewTestScope(), heartbeatGracePeriodMultiplier, maxHeartBeatInterval)

			reservation, err := m.GetOrExtendReservation(ctx, tc.mockRequest, now)
			if tc.expectError {
				assert.Error(t, err)
				assert.EqualValues(t, tc.expectedErrorStatusCode, status.Code(err))
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reservation)
			}

			if tc.expectCreateCalled {
				mockReservationStoreClient.AssertCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			} else {
				mockReservationStoreClient.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
			}

			if tc.expectUpdateCalled {
				mockReservationStoreClient.AssertCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
			} else {
				mockReservationStoreClient.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

func TestCacheManager_ReleaseReservation(t *testing.T) {
	ctx := context.Background()

	sampleOwner := "owner"

	testCases := []struct {
		name                    string
		mockRequest             *cacheservice.ReleaseReservationRequest
		expectError             bool
		expectedErrorStatusCode codes.Code
		deleteError             error
	}{
		{
			name:                    "failed validation",
			mockRequest:             &cacheservice.ReleaseReservationRequest{},
			expectError:             true,
			expectedErrorStatusCode: codes.InvalidArgument,
		},
		{
			name: "success",
			mockRequest: &cacheservice.ReleaseReservationRequest{
				Key:     sampleKey,
				OwnerId: sampleOwner,
			},
			deleteError: nil,
			expectError: false,
		},
		{
			name: "not found, success",
			mockRequest: &cacheservice.ReleaseReservationRequest{
				Key:     sampleKey,
				OwnerId: sampleOwner,
			},
			deleteError: errors.NewNotFoundError("reservation", sampleKey),
			expectError: false,
		},
		{
			name: "delete error, error",
			mockRequest: &cacheservice.ReleaseReservationRequest{
				Key:     sampleKey,
				OwnerId: sampleOwner,
			},
			deleteError:             errors.NewCacheServiceErrorf(codes.Internal, "delete error"),
			expectError:             true,
			expectedErrorStatusCode: codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestKey := fmt.Sprintf("%s:%s", "reservation", tc.mockRequest.Key)

			mockReservationStoreClient := &repoMocks.ReservationRepo{}
			mockReservationStoreClient.OnDeleteMatch(
				ctx,
				mock.MatchedBy(func(key string) bool {
					assert.Equal(t, requestKey, key)
					return true
				}), mock.MatchedBy(func(ownerId string) bool {
					assert.Equal(t, sampleOwner, ownerId)
					return true
				}),
			).Return(tc.deleteError)

			m := NewCacheManager(&mocks.CacheOutputBlobStore{}, &repoMocks.CachedOutputRepo{}, mockReservationStoreClient, 0, mockScope.NewTestScope(), time.Duration(1), time.Duration(1))

			reservation, err := m.ReleaseReservation(ctx, tc.mockRequest)
			if tc.expectError {
				assert.Error(t, err)
				assert.EqualValues(t, tc.expectedErrorStatusCode, status.Code(err))
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reservation)
			}
		})
	}
}
