package cacheservice

import (
	"context"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/cacheservice"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	catalogIdl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flytestdlib/grpcutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var (
	_ catalog.Client = &CacheClient{}
)

type CacheClient struct {
	client      cacheservice.CacheServiceClient
	store       *storage.DataStore
	maxCacheAge time.Duration
	inlineCache bool
	lruMap      *lru.Cache
}

func (c *CacheClient) GetOrExtendReservation(ctx context.Context, key catalog.Key, ownerID string, heartbeatInterval time.Duration) (*catalogIdl.Reservation, error) {
	cacheKey, err := GenerateCacheKey(ctx, key)
	if err != nil {
		logger.Errorf(ctx, "Failed to generate cache key for %v, err %v", key.Identifier, err)
		return nil, errors.Wrapf(err, "Failed to generate cache key for %v", key.Identifier.String())
	}

	response, err := c.client.GetOrExtendReservation(ctx, &cacheservice.GetOrExtendReservationRequest{
		Key:               cacheKey,
		OwnerId:           ownerID,
		HeartbeatInterval: ptypes.DurationProto(heartbeatInterval),
	})
	if err != nil {
		return nil, err
	}

	catalogReservation := &catalogIdl.Reservation{
		OwnerId:           response.Reservation.OwnerId,
		ExpiresAt:         response.Reservation.ExpiresAt,
		HeartbeatInterval: response.Reservation.HeartbeatInterval,
	}

	return catalogReservation, nil
}

func (c *CacheClient) ReleaseReservation(ctx context.Context, key catalog.Key, ownerID string) error {
	cacheKey, err := GenerateCacheKey(ctx, key)
	if err != nil {
		logger.Errorf(ctx, "Failed to generate cache key for %v, err %v", key.Identifier, err)
		return errors.Wrapf(err, "Failed to generate cache key for %v", key.Identifier.String())
	}

	_, err = c.client.ReleaseReservation(ctx, &cacheservice.ReleaseReservationRequest{
		Key:     cacheKey,
		OwnerId: ownerID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *CacheClient) Get(ctx context.Context, key catalog.Key) (catalog.Entry, error) {
	cacheKey, err := GenerateCacheKey(ctx, key)
	if err != nil {
		logger.Errorf(ctx, "Failed to generate cache key for %v, err %v", key.Identifier, err)
		return catalog.Entry{}, errors.Wrapf(err, "Failed to generate cache key for %v", key.Identifier.String())
	}

	resp, err := c.client.Get(ctx, &cacheservice.GetCacheRequest{
		Key: cacheKey,
	})
	if err != nil {
		logger.Debugf(ctx, "CacheService failed to get output for ID %s, err: %+v", key.Identifier.String(), err)
		return catalog.Entry{}, errors.Wrapf(err, "CacheService failed to get output for ID %s", key.Identifier.String())
	}

	// validate response
	if resp.GetOutput() == nil || resp.GetOutput().Output == nil || resp.GetOutput().GetMetadata() == nil {
		logger.Errorf(ctx, "Received malformed response from cache service")
		return catalog.Entry{}, status.Error(codes.Internal, "Received malformed response from cache service")
	}

	if c.maxCacheAge > time.Duration(0) {
		if time.Since(resp.GetOutput().GetMetadata().GetLastUpdatedAt().AsTime()) > c.maxCacheAge {
			logger.Infof(ctx, "Expired Cached Output %v updated on %v, older than max age %v",
				resp.GetOutput().GetMetadata().GetSourceIdentifier(), resp.GetOutput().GetMetadata().GetLastUpdatedAt(), c.maxCacheAge)
			return catalog.Entry{}, status.Error(codes.NotFound, "Artifact over age limit")
		}
	}

	outputs := &core.LiteralMap{}
	switch output := resp.Output.Output.(type) {
	case *cacheservice.CachedOutput_OutputLiterals:
		outputs = output.OutputLiterals
	case *cacheservice.CachedOutput_OutputUri:
		err = c.store.ReadProtobuf(ctx, storage.DataReference(resp.GetOutput().GetOutputUri()), outputs)
		if err != nil {
			return catalog.Entry{}, errors.Wrapf(err, "Failed to read output data from '%s'", storage.DataReference(resp.GetOutput().GetOutputUri()))
		}
	default:
		// should never happen
		return catalog.Entry{}, status.Error(codes.Internal, "Received malformed response from cache service")
	}

	source, err := GetSourceFromMetadata(resp.Output.Metadata)
	if err != nil {
		return catalog.Entry{}, errors.Wrapf(err, "failed to get source from output metadata")
	}
	if source == nil {
		return catalog.Entry{}, errors.New("failed to get source from output metadata")
	}
	md := GenerateCatalogMetadata(source, resp.Output.Metadata)

	return catalog.NewCatalogEntry(ioutils.NewInMemoryOutputReader(outputs, nil, nil), catalog.NewStatus(core.CatalogCacheStatus_CACHE_HIT, md)), nil
}

func (c *CacheClient) put(ctx context.Context, key catalog.Key, reader io.OutputReader, metadata catalog.Metadata, overwrite bool) (catalog.Status, error) {
	cacheKey, err := GenerateCacheKey(ctx, key)
	if err != nil {
		logger.Errorf(ctx, "Failed to generate cache key for %v, err %v", key.Identifier, err)
		return catalog.Status{}, errors.Wrapf(err, "Failed to generate cache key for %v", key.Identifier.String())
	}

	cacheMetadata := GenerateCacheMetadata(key, metadata)
	var cacheRequest *cacheservice.PutCacheRequest
	if c.inlineCache {
		outputs, executionErr, err := reader.Read(ctx)
		if executionErr != nil {
			logger.Errorf(ctx, "Failed to read output for %v, err %v", key.Identifier, executionErr)
			return catalog.Status{}, errors.New(fmt.Sprintf("Failed to read output for %v, err %v", key.Identifier, executionErr))
		}
		if err != nil {
			logger.Errorf(ctx, "Failed to read output for %v, err %v", key.Identifier, err)
			return catalog.Status{}, err
		}
		cacheRequest = &cacheservice.PutCacheRequest{
			Key: cacheKey,
			Output: &cacheservice.CachedOutput{
				Output: &cacheservice.CachedOutput_OutputLiterals{
					OutputLiterals: outputs,
				},
				Metadata: cacheMetadata,
			},
			Overwrite: &cacheservice.OverwriteOutput{
				Overwrite:  overwrite,
				DeleteBlob: false,
				MaxAge:     durationpb.New(c.maxCacheAge),
			},
		}
	} else {
		remoteFileOutputReader, ok := reader.(ioutils.RemoteFileOutputReader)
		if !ok {
			logger.Warnf(ctx, "Remote file output reader is expected for non-inline caching")
			return catalog.Status{}, errors.New("remote file output reader is expected for non-inline caching")
		}
		exists, err := remoteFileOutputReader.Exists(ctx)
		if err != nil {
			logger.Errorf(ctx, "Failed to check if output file exists for %v, err %v", key.Identifier, err)
			return catalog.Status{}, errors.Wrapf(err, "failed to check if output file exists for %v", key.Identifier.String())
		}
		if !exists {
			logger.Errorf(ctx, "Output file does not exist for %v", key.Identifier)
			return catalog.Status{}, status.Errorf(codes.NotFound, "Output file does not exist for %v", key.Identifier)
		}
		outputURI := remoteFileOutputReader.OutPath.GetOutputPath()

		cacheRequest = &cacheservice.PutCacheRequest{
			Key: cacheKey,
			Output: &cacheservice.CachedOutput{
				Output: &cacheservice.CachedOutput_OutputUri{
					OutputUri: outputURI.String(),
				},
				Metadata: cacheMetadata,
			},
			Overwrite: &cacheservice.OverwriteOutput{
				Overwrite:  overwrite,
				DeleteBlob: false,
				MaxAge:     durationpb.New(c.maxCacheAge),
			},
		}
	}

	_, err = c.client.Put(ctx, cacheRequest)
	if err != nil {
		logger.Errorf(ctx, "Caching output for %v returned err %v", key.Identifier, err)
		return catalog.Status{}, err
	}

	md := &core.CatalogMetadata{
		DatasetId: &key.Identifier,
	}

	return catalog.NewStatus(core.CatalogCacheStatus_CACHE_POPULATED, md), nil
}

func (c *CacheClient) Update(ctx context.Context, key catalog.Key, reader io.OutputReader, metadata catalog.Metadata) (catalog.Status, error) {
	return c.put(ctx, key, reader, metadata, true)
}

func (c *CacheClient) Put(ctx context.Context, key catalog.Key, reader io.OutputReader, metadata catalog.Metadata) (catalog.Status, error) {
	return c.put(ctx, key, reader, metadata, false)
}

func (c *CacheClient) GetReservationCache(ownerID string) catalog.ReservationCache {
	if val, ok := c.lruMap.Get(ownerID); ok {
		return val.(catalog.ReservationCache)
	}

	return catalog.ReservationCache{}
}

func (c *CacheClient) UpdateReservationCache(ownerID string, entry catalog.ReservationCache) {
	c.lruMap.Add(ownerID, entry)
}

func NewCacheClient(ctx context.Context, dataStore *storage.DataStore, endpoint string, insecureConnection bool, maxCacheAge time.Duration,
	useAdminAuth bool, maxRetries uint, backoffScalar int, backoffJitter float64, inlineCache bool, defaultServiceConfig string,
	reservationMaxCacheSize int, authOpt ...grpc.DialOption) (*CacheClient, error) {
	var opts []grpc.DialOption
	if useAdminAuth && authOpt != nil {
		opts = append(opts, authOpt...)
	}

	grpcOptions := []grpcRetry.CallOption{
		grpcRetry.WithBackoff(grpcRetry.BackoffExponentialWithJitter(time.Duration(backoffScalar)*time.Millisecond, backoffJitter)),
		grpcRetry.WithCodes(codes.DeadlineExceeded, codes.Unavailable, codes.Canceled),
		grpcRetry.WithMax(maxRetries),
	}

	if insecureConnection {
		logger.Debug(ctx, "Establishing insecure connection to CacheService")
		opts = append(opts, grpc.WithInsecure())
	} else {
		logger.Debug(ctx, "Establishing secure connection to CacheService")
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}

		creds := credentials.NewClientTLSFromCert(pool, "")
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	if defaultServiceConfig != "" {
		opts = append(opts, grpc.WithDefaultServiceConfig(defaultServiceConfig))
	}

	retryInterceptor := grpcRetry.UnaryClientInterceptor(grpcOptions...)

	tracerProvider := otelutils.GetTracerProvider(otelutils.CacheServiceClientTracer)
	opts = append(opts, grpc.WithChainUnaryInterceptor(
		grpcutils.GrpcClientMetrics().UnaryClientInterceptor(),
		otelgrpc.UnaryClientInterceptor(
			otelgrpc.WithTracerProvider(tracerProvider),
			otelgrpc.WithPropagators(propagation.TraceContext{}),
		),
		retryInterceptor))
	clientConn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return nil, err
	}
	client := cacheservice.NewCacheServiceClient(clientConn)

	var evictionFunction func(key interface{}, value interface{})
	lruCache, err := lru.NewWithEvict(reservationMaxCacheSize, evictionFunction)
	if err != nil {
		return nil, err
	}

	return &CacheClient{
		client:      client,
		store:       dataStore,
		maxCacheAge: maxCacheAge,
		inlineCache: inlineCache,
		lruMap:      lruCache,
	}, nil

}
