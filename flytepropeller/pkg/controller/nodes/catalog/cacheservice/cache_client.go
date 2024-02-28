package cacheservice

import (
	"context"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/cacheservice"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	catalogIdl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
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
}

func (c *CacheClient) GetOrExtendReservation(ctx context.Context, key catalog.Key, ownerID string, heartbeatInterval time.Duration) (*catalogIdl.Reservation, error) {
	cacheKey, err := GenerateCacheKey(ctx, key)
	if err != nil {
		logger.Debugf(ctx, "Failed to generate cache key for %v, err %v", key.Identifier, err)
		return nil, err
	}

	response, err := c.client.GetOrExtendReservation(ctx, &cacheservice.GetOrExtendReservationRequest{
		Key:               cacheKey,
		OwnerId:           ownerID,
		HeartbeatInterval: ptypes.DurationProto(heartbeatInterval),
	})
	if err != nil {
		logger.Debugf(ctx, "CacheService failed to get or extend reservation for ID %s, err: %+v", key.Identifier.String(), err)
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
		logger.Debugf(ctx, "Failed to generate cache key for %v, err %v", key.Identifier, err)
		return err
	}

	_, err = c.client.ReleaseReservation(ctx, &cacheservice.ReleaseReservationRequest{
		Key:     cacheKey,
		OwnerId: ownerID,
	})
	if err != nil {
		logger.Debugf(ctx, "CacheService failed to release reservation for ID %s, err: %+v", key.Identifier.String(), err)
		return err
	}
	return nil
}

func (c *CacheClient) Get(ctx context.Context, key catalog.Key) (catalog.Entry, error) {
	cacheKey, err := GenerateCacheKey(ctx, key)
	if err != nil {
		logger.Debugf(ctx, "Failed to generate cache key for %v, err %v", key.Identifier, err)
		return catalog.Entry{}, err
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
		return catalog.Entry{}, status.Error(codes.Internal, "Received malformed response from cache service")
	}

	if c.maxCacheAge > time.Duration(0) {
		if time.Since(resp.GetOutput().GetMetadata().CreatedAt.AsTime()) > c.maxCacheAge {
			logger.Warningf(ctx, "Expired Cached Output %v created on %v, older than max age %v",
				resp.GetOutput().GetMetadata().GetSourceIdentifier(), resp.GetOutput().GetMetadata().GetCreatedAt(), c.maxCacheAge)
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
			logger.Errorf(ctx, "Failed to retrieve output data from '%s' with error '%s'", storage.DataReference(resp.GetOutput().GetOutputUri()), err)
			return catalog.Entry{}, err
		}
	default:
		// should never happen
		return catalog.Entry{}, status.Error(codes.Internal, "Received malformed response from cache service")
	}

	source, err := GetSourceFromMetadata(resp.Output.Metadata)
	if err != nil {
		return catalog.Entry{}, errors.Wrapf(err, "failed to get source from metadata")
	}
	if source == nil {
		return catalog.Entry{}, errors.New("failed to get source from metadata")
	}
	md := GenerateCatalogMetadata(source, resp.Output.Metadata)

	return catalog.NewCatalogEntry(ioutils.NewInMemoryOutputReader(outputs, nil, nil), catalog.NewStatus(core.CatalogCacheStatus_CACHE_HIT, md)), nil
}

func (c *CacheClient) put(ctx context.Context, key catalog.Key, reader io.OutputReader, metadata catalog.Metadata, overwrite bool) (catalog.Status, error) {
	outputs, executionErr, err := reader.Read(ctx)
	if executionErr != nil {
		logger.Debugf(ctx, "Failed to read output for %v, err %v", key.Identifier, executionErr)
		return catalog.Status{}, errors.New(fmt.Sprintf("Failed to read output for %v, err %v", key.Identifier, executionErr))
	}
	if err != nil {
		logger.Debugf(ctx, "Failed to read output for %v, err %v", key.Identifier, err)
		return catalog.Status{}, err
	}
	cacheKey, err := GenerateCacheKey(ctx, key)
	if err != nil {
		logger.Debugf(ctx, "Failed to generate cache key for %v, err %v", key.Identifier, err)
		return catalog.Status{}, err
	}
	cacheRequest := &cacheservice.PutCacheRequest{
		Key: cacheKey,
		Output: &cacheservice.CachedOutput{
			Output: &cacheservice.CachedOutput_OutputLiterals{
				OutputLiterals: outputs,
			},
			Metadata: GenerateCacheMetadata(key, metadata),
		},
		Overwrite: overwrite,
	}

	_, err = c.client.Put(ctx, cacheRequest)
	if err != nil {
		logger.Debugf(ctx, "Caching output for %v returned err %v", key.Identifier, err)
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

func NewCacheClient(ctx context.Context, dataStore *storage.DataStore, endpoint string, insecureConnection bool, maxCacheAge time.Duration,
	useAdminAuth bool, maxRetries int, maxPerRetryTimeout time.Duration, backOffScalar int, defaultServiceConfig string, authOpt ...grpc.DialOption) (*CacheClient, error) {
	var opts []grpc.DialOption
	if useAdminAuth && authOpt != nil {
		opts = append(opts, authOpt...)
	}

	grpcOptions := []grpcRetry.CallOption{
		grpcRetry.WithBackoff(grpcRetry.BackoffLinear(time.Duration(backOffScalar) * time.Millisecond)),
		grpcRetry.WithCodes(codes.DeadlineExceeded, codes.Unavailable, codes.Canceled),
		grpcRetry.WithMax(uint(maxRetries)),
		grpcRetry.WithPerRetryTimeout(maxPerRetryTimeout),
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
		grpcPrometheus.UnaryClientInterceptor,
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

	return &CacheClient{
		client:      client,
		store:       dataStore,
		maxCacheAge: maxCacheAge,
	}, nil

}
