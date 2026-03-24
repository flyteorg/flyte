package cache_service

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	catalog "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/catalog"
	ioMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/ioutils"
	contextutils "github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils/labeled"
	mem "github.com/flyteorg/flyte/v2/flytestdlib/storage"
	cacheservicepb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice"
	cacheservicev2 "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice/v2"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice/v2/v2connect"
	corepb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
}

type stubCacheService struct {
	getFunc                    func(context.Context, *connect.Request[cacheservicev2.GetCacheRequest]) (*connect.Response[cacheservicepb.GetCacheResponse], error)
	putFunc                    func(context.Context, *connect.Request[cacheservicev2.PutCacheRequest]) (*connect.Response[cacheservicepb.PutCacheResponse], error)
	getOrExtendReservationFunc func(context.Context, *connect.Request[cacheservicev2.GetOrExtendReservationRequest]) (*connect.Response[cacheservicepb.GetOrExtendReservationResponse], error)
	releaseReservationFunc     func(context.Context, *connect.Request[cacheservicev2.ReleaseReservationRequest]) (*connect.Response[cacheservicepb.ReleaseReservationResponse], error)
}

var _ v2connect.CacheServiceClient = (*stubCacheService)(nil)

func (s *stubCacheService) Get(ctx context.Context, req *connect.Request[cacheservicev2.GetCacheRequest]) (*connect.Response[cacheservicepb.GetCacheResponse], error) {
	return s.getFunc(ctx, req)
}

func (s *stubCacheService) Put(ctx context.Context, req *connect.Request[cacheservicev2.PutCacheRequest]) (*connect.Response[cacheservicepb.PutCacheResponse], error) {
	return s.putFunc(ctx, req)
}

func (s *stubCacheService) Delete(context.Context, *connect.Request[cacheservicev2.DeleteCacheRequest]) (*connect.Response[cacheservicepb.DeleteCacheResponse], error) {
	return nil, nil
}

func (s *stubCacheService) GetOrExtendReservation(ctx context.Context, req *connect.Request[cacheservicev2.GetOrExtendReservationRequest]) (*connect.Response[cacheservicepb.GetOrExtendReservationResponse], error) {
	return s.getOrExtendReservationFunc(ctx, req)
}

func (s *stubCacheService) ReleaseReservation(ctx context.Context, req *connect.Request[cacheservicev2.ReleaseReservationRequest]) (*connect.Response[cacheservicepb.ReleaseReservationResponse], error) {
	return s.releaseReservationFunc(ctx, req)
}

func TestBuildCacheKey(t *testing.T) {
	reader := ioMocks.NewInputReader(t)

	key := catalog.Key{
		Identifier: &corepb.Identifier{
			ResourceType: corepb.ResourceType_TASK,
			Project:      "project",
			Domain:       "domain",
			Name:         "task",
			Version:      "ignored",
			Org:          "org",
		},
		CacheVersion:   "cache-v1",
		TypedInterface: &corepb.TypedInterface{},
		InputReader:    reader,
	}

	cacheKey, err := buildCacheKey(context.Background(), key)
	require.NoError(t, err)
	assert.Contains(t, cacheKey, "-cache-v1")
}

func TestBuildCacheKeyIgnoresPrecomputedCacheKeyShortcut(t *testing.T) {
	reader := ioMocks.NewInputReader(t)

	key := catalog.Key{
		Identifier: &corepb.Identifier{
			ResourceType: corepb.ResourceType_TASK,
			Project:      "project",
			Domain:       "domain",
			Name:         "task",
			Version:      "ignored",
			Org:          "org",
		},
		CacheVersion:   "cache-v1",
		CacheKey:       "precomputed-key-should-not-be-used",
		TypedInterface: &corepb.TypedInterface{},
		InputReader:    reader,
	}

	cacheKey, err := buildCacheKey(context.Background(), key)
	require.NoError(t, err)
	assert.NotEqual(t, "precomputed-key-should-not-be-used", cacheKey)
	assert.Contains(t, cacheKey, "-cache-v1")
}

func TestGetMapsNotFound(t *testing.T) {
	store := newTestDataStore(t)
	client := NewWithServiceClient(&stubCacheService{
		getFunc: func(context.Context, *connect.Request[cacheservicev2.GetCacheRequest]) (*connect.Response[cacheservicepb.GetCacheResponse], error) {
			return nil, connect.NewError(connect.CodeNotFound, assert.AnError)
		},
	}, store, 0)

	_, err := client.Get(context.Background(), newCatalogKey(t))
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, grpcstatus.Code(err))
}

func TestPutUsesOutputURI(t *testing.T) {
	store := newTestDataStore(t)
	var got *cacheservicev2.PutCacheRequest

	client := NewWithServiceClient(&stubCacheService{
		putFunc: func(_ context.Context, req *connect.Request[cacheservicev2.PutCacheRequest]) (*connect.Response[cacheservicepb.PutCacheResponse], error) {
			got = req.Msg
			return connect.NewResponse(&cacheservicepb.PutCacheResponse{}), nil
		},
	}, store, 0)

	outputPrefix := mem.DataReference("s3://bucket/prefix")
	reader := ioutils.NewRemoteFileOutputReader(context.Background(), store, ioutils.NewReadOnlyOutputFilePaths(context.Background(), store, outputPrefix), 0)
	require.NoError(t, store.WriteProtobuf(context.Background(), mem.DataReference("s3://bucket/prefix/outputs.pb"), mem.Options{}, &corepb.LiteralMap{}))
	status, err := client.Put(context.Background(), newCatalogKey(t), reader, catalog.Metadata{CreatedAt: timestamppb.Now()})
	require.NoError(t, err)
	assert.Equal(t, corepb.CatalogCacheStatus_CACHE_POPULATED, status.GetCacheStatus())
	require.NotNil(t, got)
	assert.Equal(t, "s3://bucket/prefix/outputs.pb", got.GetBaseRequest().GetOutput().GetOutputUri())
}

func TestGetOrExtendReservationPassesHeartbeat(t *testing.T) {
	store := newTestDataStore(t)
	client := NewWithServiceClient(&stubCacheService{
		getOrExtendReservationFunc: func(_ context.Context, req *connect.Request[cacheservicev2.GetOrExtendReservationRequest]) (*connect.Response[cacheservicepb.GetOrExtendReservationResponse], error) {
			assert.Equal(t, "owner-1", req.Msg.GetBaseRequest().GetOwnerId())
			assert.Equal(t, 5*time.Second, req.Msg.GetBaseRequest().GetHeartbeatInterval().AsDuration())
			return connect.NewResponse(&cacheservicepb.GetOrExtendReservationResponse{
				Reservation: &cacheservicepb.Reservation{
					Key:               req.Msg.GetBaseRequest().GetKey(),
					OwnerId:           req.Msg.GetBaseRequest().GetOwnerId(),
					HeartbeatInterval: durationpb.New(5 * time.Second),
				},
			}), nil
		},
	}, store, 0)

	reservation, err := client.GetOrExtendReservation(context.Background(), newCatalogKey(t), "owner-1", 5*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "owner-1", reservation.GetOwnerId())
}

func TestReservationCache(t *testing.T) {
	client := &Client{}
	entry := catalog.ReservationCache{
		Timestamp:         time.Now(),
		ReservationStatus: corepb.CatalogReservation_RESERVATION_ACQUIRED,
	}

	client.UpdateReservationCache("owner-1", entry)
	assert.Equal(t, entry, client.GetReservationCache("owner-1"))
}

func TestGetRespectsMaxCacheAge(t *testing.T) {
	store := newTestDataStore(t)
	key := newCatalogKey(t)
	client := NewWithServiceClient(&stubCacheService{
		getFunc: func(context.Context, *connect.Request[cacheservicev2.GetCacheRequest]) (*connect.Response[cacheservicepb.GetCacheResponse], error) {
			return connect.NewResponse(&cacheservicepb.GetCacheResponse{
				Output: &cacheservicepb.CachedOutput{
					Output: &cacheservicepb.CachedOutput_OutputUri{
						OutputUri: "s3://bucket/prefix/outputs.pb",
					},
					Metadata: &cacheservicepb.Metadata{
						SourceIdentifier: key.Identifier,
						LastUpdatedAt:    timestamppb.New(time.Now().Add(-2 * time.Hour)),
					},
				},
			}), nil
		},
	}, store, time.Hour)

	_, err := client.Get(context.Background(), key)
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, grpcstatus.Code(err))
}

func newCatalogKey(t *testing.T) catalog.Key {
	reader := ioMocks.NewInputReader(t)

	return catalog.Key{
		Identifier: &corepb.Identifier{
			ResourceType: corepb.ResourceType_TASK,
			Project:      "project",
			Domain:       "domain",
			Name:         "task",
			Version:      "v1",
			Org:          "org",
		},
		CacheVersion:   "cache-v1",
		TypedInterface: &corepb.TypedInterface{},
		InputReader:    reader,
	}
}

func newTestDataStore(t *testing.T) *mem.DataStore {
	t.Helper()

	store, err := mem.NewDataStore(&mem.Config{
		Type: mem.TypeMemory,
	}, promutils.NewTestScope())
	require.NoError(t, err)
	return store
}
