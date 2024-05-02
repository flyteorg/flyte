package impl

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/cacheservice/pkg/errors"
	"github.com/flyteorg/flyte/cacheservice/pkg/manager/impl/validators"
	"github.com/flyteorg/flyte/cacheservice/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyte/cacheservice/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/cacheservice/pkg/repositories/models"
	"github.com/flyteorg/flyte/cacheservice/pkg/repositories/transformers"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/cacheservice"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

type cacheMetrics struct {
	scope                            promutils.Scope
	putResponseTime                  labeled.StopWatch
	updateResponseTime               labeled.StopWatch
	getResponseTime                  labeled.StopWatch
	deleteResponseTime               labeled.StopWatch
	getReservationResponseTime       labeled.StopWatch
	releaseReservationResponseTime   labeled.StopWatch
	validationErrorCounter           labeled.Counter
	getFailureCounter                labeled.Counter
	getSuccessCounter                labeled.Counter
	putFailureCounter                labeled.Counter
	putSuccessCounter                labeled.Counter
	updateFailureCounter             labeled.Counter
	updateSuccessCounter             labeled.Counter
	deleteFailureCounter             labeled.Counter
	deleteSuccessCounter             labeled.Counter
	createDataFailureCounter         labeled.Counter
	createDataSuccessCounter         labeled.Counter
	getReservationFailureCounter     labeled.Counter
	getReservationSuccessCounter     labeled.Counter
	releaseReservationFailureCounter labeled.Counter
	releaseReservationSuccessCounter labeled.Counter
	notFoundCounter                  labeled.Counter
	alreadyExistsCount               labeled.Counter
}

type cacheManager struct {
	outputStore                    interfaces.CacheOutputBlobStore
	dataStore                      repoInterfaces.CachedOutputRepo
	reservationStore               repoInterfaces.ReservationRepo
	systemMetrics                  cacheMetrics
	maxInlineSizeBytes             int64
	heartbeatGracePeriodMultiplier time.Duration
	maxHeartbeatInterval           time.Duration
}

// Get retrieves the cached output for the given key. If the key is not found, a NotFound error is returned.
func (m *cacheManager) Get(ctx context.Context, request *cacheservice.GetCacheRequest) (*cacheservice.GetCacheResponse, error) {
	timer := m.systemMetrics.getResponseTime.Start(ctx)
	defer timer.Stop()

	err := validators.ValidateGetCacheRequest(request)
	if err != nil {
		logger.Warningf(ctx, "Invalid get cache request %v, err: %v", request, err)
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	outputModel, err := m.dataStore.Get(ctx, request.Key)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			logger.Debugf(ctx, "Key %v not found in data store", request.Key)
			m.systemMetrics.notFoundCounter.Inc(ctx)
		} else {
			logger.Errorf(ctx, "Failed to get output in data store, err: %v", err)
			m.systemMetrics.getFailureCounter.Inc(ctx)
		}
		return nil, err
	}

	output, err := transformers.FromCachedOutputModel(ctx, outputModel)
	if err != nil {
		logger.Errorf(ctx, "Failed to transform output model to cached output, err: %v", err)
		m.systemMetrics.getFailureCounter.Inc(ctx)
		return nil, errors.NewCacheServiceErrorf(codes.Internal, "Failed to transform output model to cached output, err: %v", err)
	}

	m.systemMetrics.getSuccessCounter.Inc(ctx)
	return &cacheservice.GetCacheResponse{Output: output}, nil
}

// Put stores the given output in the cache with the given key. If the key already exists and overwrite is false, an AlreadyExists error is returned.
func (m *cacheManager) Put(ctx context.Context, request *cacheservice.PutCacheRequest) (*cacheservice.PutCacheResponse, error) {
	timer := m.systemMetrics.putResponseTime.Start(ctx)
	defer timer.Stop()

	err := validators.ValidatePutCacheRequest(request)
	if err != nil {
		logger.Warningf(ctx, "Invalid put cache request %v, err: %v", request, err)
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	cachedOutputModel, err := transformers.CreateCachedOutputModel(ctx, request.Key, request.Output)
	if err != nil {
		logger.Errorf(ctx, "Failed to create cached output model, err: %v", err)
		m.systemMetrics.putFailureCounter.Inc(ctx)
		return nil, errors.NewCacheServiceErrorf(codes.Internal, "Failed to create cached output model, err: %v", err)
	}

	// TODO - @pvditt - can do this in single transaction w/ postgres client - move logic to client (still need to delete blob)
	cachedOutput, err := m.dataStore.Get(ctx, request.Key)
	var notFound bool
	if err != nil {
		if status.Code(err) != codes.NotFound {
			logger.Errorf(ctx, "Failed to check if output with key %v is already in data store, err: %v", request.Key, err)
			return nil, err
		}
		notFound = true
	}
	if !notFound && cachedOutput != nil {
		if request.Overwrite == nil || !request.Overwrite.Overwrite {
			m.systemMetrics.alreadyExistsCount.Inc(ctx)
			logger.Errorf(ctx, "Output with key %v already exists", request.Key)
			return nil, errors.NewCacheServiceErrorf(codes.AlreadyExists, "Output with key %v already exists", request.Key)
		}
		if cachedOutput.OutputURI != "" && request.Overwrite.DeleteBlob {
			err = m.outputStore.Delete(ctx, cachedOutput.OutputURI)
			if err != nil {
				logger.Errorf(ctx, "Failed to delete output in blob store before overwriting, err: %v", err)
				return nil, err
			}
		}
	}

	switch output := request.Output.Output.(type) {
	case *cacheservice.CachedOutput_OutputLiterals:
		if m.maxInlineSizeBytes == 0 || int64(proto.Size(output.OutputLiterals)) > m.maxInlineSizeBytes {
			dataLocation, err := m.outputStore.Create(ctx, request.GetKey(), output.OutputLiterals)
			if err != nil {
				logger.Errorf(ctx, "Failed to put output in blob store, err: %v", err)
				m.systemMetrics.createDataFailureCounter.Inc(ctx)
				return nil, err
			}
			cachedOutputModel.OutputLiteral = nil
			cachedOutputModel.OutputURI = dataLocation
		}
		err := m.dataStore.Put(ctx, request.Key, cachedOutputModel)
		if err != nil {
			logger.Errorf(ctx, "Failed to put output in data store, err: %v", err)
			m.systemMetrics.putFailureCounter.Inc(ctx)
			return nil, err
		}
	case *cacheservice.CachedOutput_OutputUri:
		err := m.dataStore.Put(ctx, request.Key, cachedOutputModel)
		if err != nil {
			logger.Errorf(ctx, "Failed to put output uri in data store, err: %v", err)
			m.systemMetrics.putFailureCounter.Inc(ctx)
			return nil, err
		}
	default:
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, errors.NewCacheServiceErrorf(codes.InvalidArgument, "Invalid output type %v", output)
	}

	m.systemMetrics.putSuccessCounter.Inc(ctx)
	return &cacheservice.PutCacheResponse{}, nil
}

func (m *cacheManager) Delete(ctx context.Context, request *cacheservice.DeleteCacheRequest) (*cacheservice.DeleteCacheResponse, error) {
	timer := m.systemMetrics.deleteResponseTime.Start(ctx)
	defer timer.Stop()

	return nil, errors.NewCacheServiceErrorf(codes.Unimplemented, "Delete endpoint not implemented")
}

// GetOrExtendReservation retrieves the reservation for the given key. If the reservation does not exist, a new reservation is created.
// If the reservation exists and is held by a different owner and not expired, the reservation is not extended and the existing reservation is returned.
func (m *cacheManager) GetOrExtendReservation(ctx context.Context, request *cacheservice.GetOrExtendReservationRequest, now time.Time) (*cacheservice.GetOrExtendReservationResponse, error) {
	timer := m.systemMetrics.getReservationResponseTime.Start(ctx)
	defer timer.Stop()

	err := validators.ValidateGetOrExtendReservationRequest(request)
	if err != nil {
		logger.Warningf(ctx, "Invalid GetOrExtendReservation request %v, err: %v", request, err)
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	resKey := fmt.Sprintf("%s:%s", "reservation", request.Key)

	reservationModel, err := m.reservationStore.Get(ctx, resKey)
	reservationExists := true
	if err != nil {
		if status.Code(err) == codes.NotFound {
			reservationExists = false
		} else {
			logger.Errorf(ctx, "Failed to Get reservation in reservation store, err: %v", err)
			m.systemMetrics.getReservationFailureCounter.Inc(ctx)
			return nil, err
		}
	}

	heartbeatInterval := m.maxHeartbeatInterval
	if request.GetHeartbeatInterval() != nil && request.GetHeartbeatInterval().AsDuration() < m.maxHeartbeatInterval {
		heartbeatInterval = request.GetHeartbeatInterval().AsDuration()
	}

	newReservation := &models.CacheReservation{
		Key:       resKey,
		OwnerID:   request.OwnerId,
		ExpiresAt: now.Add(heartbeatInterval * m.heartbeatGracePeriodMultiplier),
	}

	var storeError error
	if reservationExists {
		if reservationModel.ExpiresAt.Before(now) || reservationModel.OwnerID == request.OwnerId {
			storeError = m.reservationStore.Update(ctx, newReservation, now)
		} else {
			logger.Debugf(ctx, "CacheReservation: %+v is held by %s", reservationModel.Key, reservationModel.OwnerID)
			reservation := transformers.FromReservationModel(ctx, reservationModel)
			return &cacheservice.GetOrExtendReservationResponse{Reservation: reservation}, nil
		}
	} else {
		storeError = m.reservationStore.Create(ctx, newReservation, now)
	}

	if storeError != nil {
		if status.Code(storeError) == codes.AlreadyExists {
			logger.Debugf(ctx, "CacheReservation: %+v already exists", newReservation.Key)
			newReservation, err = m.reservationStore.Get(ctx, resKey)
			if err != nil {
				logger.Errorf(ctx, "Failed to Get reservation in reservation store, err: %v", err)
				m.systemMetrics.getReservationFailureCounter.Inc(ctx)
				return nil, err
			}
		} else {
			logger.Errorf(ctx, "Failed to Create/Update reservation in reservation store, err: %v", storeError)
			m.systemMetrics.getReservationFailureCounter.Inc(ctx)
			return nil, storeError
		}
	}

	reservation := transformers.FromReservationModel(ctx, newReservation)
	m.systemMetrics.getReservationSuccessCounter.Inc(ctx)
	return &cacheservice.GetOrExtendReservationResponse{Reservation: reservation}, nil
}

// ReleaseReservation releases the reservation for the given key and owner. If the reservation does not exist return gracefully
func (m *cacheManager) ReleaseReservation(ctx context.Context, request *cacheservice.ReleaseReservationRequest) (*cacheservice.ReleaseReservationResponse, error) {
	timer := m.systemMetrics.releaseReservationResponseTime.Start(ctx)
	defer timer.Stop()

	err := validators.ValidateReleaseReservationRequest(request)
	if err != nil {
		logger.Warningf(ctx, "Invalid ReleaseReservation request %v, err: %v", request, err)
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	resKey := fmt.Sprintf("%s:%s", "reservation", request.Key)

	err = m.reservationStore.Delete(ctx, resKey, request.OwnerId)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			logger.Debugf(ctx, "CacheReservation with key %v and owner %v not found", request.Key, request.OwnerId)
			m.systemMetrics.notFoundCounter.Inc(ctx)
			return &cacheservice.ReleaseReservationResponse{}, nil
		}
		m.systemMetrics.releaseReservationFailureCounter.Inc(ctx)
		logger.Errorf(ctx, "Failed to Release reservation in reservation store, err: %v", err)
		return nil, err
	}

	m.systemMetrics.releaseReservationSuccessCounter.Inc(ctx)
	return &cacheservice.ReleaseReservationResponse{}, nil
}

func NewCacheManager(outputStore interfaces.CacheOutputBlobStore, dataStore repoInterfaces.CachedOutputRepo, reservationStore repoInterfaces.ReservationRepo, maxInlineSizeBytes int64, cacheScope promutils.Scope,
	heartbeatGracePeriodMultiplier time.Duration, maxHeartbeatInterval time.Duration) interfaces.CacheManager {
	cacheMetrics := cacheMetrics{
		scope:                            cacheScope,
		putResponseTime:                  labeled.NewStopWatch("put_duration", "The duration of creating new cached outputs", time.Millisecond, cacheScope, labeled.EmitUnlabeledMetric),
		updateResponseTime:               labeled.NewStopWatch("update_duration", "The duration of updating cached outputs", time.Millisecond, cacheScope, labeled.EmitUnlabeledMetric),
		getResponseTime:                  labeled.NewStopWatch("get_duration", "The duration of fetching cached outputs", time.Millisecond, cacheScope, labeled.EmitUnlabeledMetric),
		deleteResponseTime:               labeled.NewStopWatch("delete_duration", "The duration of deleting cached outputs", time.Millisecond, cacheScope, labeled.EmitUnlabeledMetric),
		getReservationResponseTime:       labeled.NewStopWatch("get_reservation_duration", "The duration of getting or extending reservations", time.Millisecond, cacheScope, labeled.EmitUnlabeledMetric),
		releaseReservationResponseTime:   labeled.NewStopWatch("release_reservation_duration", "The duration of releasing reservations", time.Millisecond, cacheScope, labeled.EmitUnlabeledMetric),
		validationErrorCounter:           labeled.NewCounter("validation_error", "The number of cache service validation errors.", cacheScope, labeled.EmitUnlabeledMetric),
		getFailureCounter:                labeled.NewCounter("get_failure", "The number of cache get failures.", cacheScope, labeled.EmitUnlabeledMetric),
		getSuccessCounter:                labeled.NewCounter("get_success", "The number of cache get successes.", cacheScope, labeled.EmitUnlabeledMetric),
		putFailureCounter:                labeled.NewCounter("put_failure", "The number of cache put failures.", cacheScope, labeled.EmitUnlabeledMetric),
		putSuccessCounter:                labeled.NewCounter("put_success", "The number of cache put successes.", cacheScope, labeled.EmitUnlabeledMetric),
		updateFailureCounter:             labeled.NewCounter("update_failure", "The number of cache update failures.", cacheScope, labeled.EmitUnlabeledMetric),
		updateSuccessCounter:             labeled.NewCounter("update_success", "The number of cache update successes.", cacheScope, labeled.EmitUnlabeledMetric),
		deleteFailureCounter:             labeled.NewCounter("delete_failure", "The number of cache delete failures.", cacheScope, labeled.EmitUnlabeledMetric),
		deleteSuccessCounter:             labeled.NewCounter("delete_success", "The number of cache delete successes.", cacheScope, labeled.EmitUnlabeledMetric),
		createDataFailureCounter:         labeled.NewCounter("create_data_failure", "The number of create blob data failures.", cacheScope, labeled.EmitUnlabeledMetric),
		getReservationSuccessCounter:     labeled.NewCounter("get_reservation_success", "The number of get reservation successes.", cacheScope, labeled.EmitUnlabeledMetric),
		createDataSuccessCounter:         labeled.NewCounter("create_data_success", "The number of create blob data successes.", cacheScope, labeled.EmitUnlabeledMetric),
		getReservationFailureCounter:     labeled.NewCounter("get_reservation_failure", "The number of get reservation failures.", cacheScope, labeled.EmitUnlabeledMetric),
		releaseReservationSuccessCounter: labeled.NewCounter("release_reservation_success", "The number of release reservation successes.", cacheScope, labeled.EmitUnlabeledMetric),
		releaseReservationFailureCounter: labeled.NewCounter("release_reservation_failure", "The number of release reservation failures.", cacheScope, labeled.EmitUnlabeledMetric),
		notFoundCounter:                  labeled.NewCounter("not_found", "The number of cache keys not found.", cacheScope, labeled.EmitUnlabeledMetric),
		alreadyExistsCount:               labeled.NewCounter("already_exists", "The number of cache keys already exists.", cacheScope, labeled.EmitUnlabeledMetric),
	}

	return &cacheManager{
		outputStore:                    outputStore,
		dataStore:                      dataStore,
		reservationStore:               reservationStore,
		systemMetrics:                  cacheMetrics,
		maxInlineSizeBytes:             maxInlineSizeBytes,
		heartbeatGracePeriodMultiplier: heartbeatGracePeriodMultiplier,
		maxHeartbeatInterval:           maxHeartbeatInterval,
	}
}
