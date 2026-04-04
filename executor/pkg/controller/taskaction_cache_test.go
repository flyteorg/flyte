package controller

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	executorplugin "github.com/flyteorg/flyte/v2/executor/pkg/plugin"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/catalog"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice"
	corepb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type stubCatalogClient struct {
	getFunc                    func(context.Context, catalog.Key) (catalog.Entry, error)
	getOrExtendReservationFunc func(context.Context, catalog.Key, string, time.Duration) (*cacheservice.Reservation, error)
	putFunc                    func(context.Context, catalog.Key, io.OutputReader, catalog.Metadata) (catalog.Status, error)
	releaseReservationFunc     func(context.Context, catalog.Key, string) error
}

func (s *stubCatalogClient) Get(ctx context.Context, key catalog.Key) (catalog.Entry, error) {
	return s.getFunc(ctx, key)
}

func (s *stubCatalogClient) GetOrExtendReservation(ctx context.Context, key catalog.Key, ownerID string, heartbeatInterval time.Duration) (*cacheservice.Reservation, error) {
	return s.getOrExtendReservationFunc(ctx, key, ownerID, heartbeatInterval)
}

func (s *stubCatalogClient) Put(ctx context.Context, key catalog.Key, reader io.OutputReader, metadata catalog.Metadata) (catalog.Status, error) {
	return s.putFunc(ctx, key, reader, metadata)
}

func (s *stubCatalogClient) Update(context.Context, catalog.Key, io.OutputReader, catalog.Metadata) (catalog.Status, error) {
	return catalog.Status{}, nil
}

func (s *stubCatalogClient) ReleaseReservation(ctx context.Context, key catalog.Key, ownerID string) error {
	return s.releaseReservationFunc(ctx, key, ownerID)
}

func (s *stubCatalogClient) GetReservationCache(string) catalog.ReservationCache {
	return catalog.ReservationCache{}
}

func (s *stubCatalogClient) UpdateReservationCache(string, catalog.ReservationCache) {}

func TestBuildTaskCacheConfig(t *testing.T) {
	ensureTestMetricKeys()
	ctx := context.Background()
	taskAction, dataStore := newCacheableTaskAction(t, true, true)
	tCtx := newTaskExecutionContext(t, taskAction, dataStore)

	cfg, ok, err := buildTaskCacheConfig(ctx, taskAction, tCtx)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Empty(t, cfg.key.CacheKey)
	assert.Equal(t, taskAction.Spec.CacheKey, cfg.key.CacheVersion)
	assert.Equal(t, []string{"ignored_input"}, cfg.key.CacheIgnoreInputVars)
	assert.True(t, cfg.serializable)
	assert.Equal(t, "default/cacheable-action", cfg.ownerID)
}

func TestHandleCacheBeforeExecutionHit(t *testing.T) {
	ensureTestMetricKeys()
	ctx := context.Background()
	taskAction, dataStore := newCacheableTaskAction(t, true, false)
	tCtx := newTaskExecutionContext(t, taskAction, dataStore)
	expected := &corepb.LiteralMap{
		Literals: map[string]*corepb.Literal{
			"o0": {
				Value: &corepb.Literal_Scalar{
					Scalar: &corepb.Scalar{
						Value: &corepb.Scalar_Primitive{
							Primitive: &corepb.Primitive{Value: &corepb.Primitive_StringValue{StringValue: "cached"}},
						},
					},
				},
			},
		},
	}

	r := &TaskActionReconciler{
		DataStore: dataStore,
		Catalog: &stubCatalogClient{
			getFunc: func(context.Context, catalog.Key) (catalog.Entry, error) {
				return catalog.NewCatalogEntry(ioutils.NewInMemoryOutputReader(expected, nil, nil), catalog.NewStatus(corepb.CatalogCacheStatus_CACHE_HIT, nil)), nil
			},
		},
	}

	transition, handled, err := r.evaluateCacheBeforeExecution(ctx, taskAction, tCtx)
	require.NoError(t, err)
	require.True(t, handled)
	assert.Equal(t, pluginsCore.PhaseSuccess, transition.Info().Phase())
	assert.Equal(t, corepb.CatalogCacheStatus_CACHE_HIT, transition.Info().Info().ExternalResources[0].CacheStatus)

	actual := &corepb.LiteralMap{}
	require.NoError(t, dataStore.ReadProtobuf(ctx, tCtx.OutputWriter().GetOutputPath(), actual))
	assert.True(t, proto.Equal(expected, actual))
}

func TestHandleCacheBeforeExecutionWaitingForReservation(t *testing.T) {
	ensureTestMetricKeys()
	ctx := context.Background()
	taskAction, dataStore := newCacheableTaskAction(t, true, true)
	tCtx := newTaskExecutionContext(t, taskAction, dataStore)

	r := &TaskActionReconciler{
		DataStore: dataStore,
		Catalog: &stubCatalogClient{
			getFunc: func(context.Context, catalog.Key) (catalog.Entry, error) {
				return catalog.Entry{}, grpcstatus.Error(codes.NotFound, "miss")
			},
			getOrExtendReservationFunc: func(_ context.Context, _ catalog.Key, ownerID string, heartbeat time.Duration) (*cacheservice.Reservation, error) {
				assert.Equal(t, cacheReservationHeartbeatInterval, heartbeat)
				assert.Equal(t, "default/cacheable-action", ownerID)
				return &cacheservice.Reservation{OwnerId: "default/other-action"}, nil
			},
		},
	}

	transition, handled, err := r.evaluateCacheBeforeExecution(ctx, taskAction, tCtx)
	require.NoError(t, err)
	require.True(t, handled)
	assert.Equal(t, pluginsCore.PhaseWaitingForCache, transition.Info().Phase())
	assert.Equal(t, corepb.CatalogCacheStatus_CACHE_MISS, transition.Info().Info().ExternalResources[0].CacheStatus)
}

func TestHandleCacheBeforeExecutionMissWithoutSerializableSkipsReservation(t *testing.T) {
	ensureTestMetricKeys()
	ctx := context.Background()
	taskAction, dataStore := newCacheableTaskAction(t, true, false)
	tCtx := newTaskExecutionContext(t, taskAction, dataStore)

	r := &TaskActionReconciler{
		DataStore: dataStore,
		Catalog: &stubCatalogClient{
			getFunc: func(context.Context, catalog.Key) (catalog.Entry, error) {
				return catalog.Entry{}, grpcstatus.Error(codes.NotFound, "miss")
			},
			getOrExtendReservationFunc: func(context.Context, catalog.Key, string, time.Duration) (*cacheservice.Reservation, error) {
				t.Fatal("reservation should not be requested for non-serializable cache")
				return nil, nil
			},
		},
	}

	transition, handled, err := r.evaluateCacheBeforeExecution(ctx, taskAction, tCtx)
	require.NoError(t, err)
	require.False(t, handled)
	assert.Equal(t, pluginsCore.UnknownTransition, transition)
}

func TestHandleCacheAfterExecutionWritesBackAndReleasesReservation(t *testing.T) {
	ensureTestMetricKeys()
	ctx := context.Background()
	taskAction, dataStore := newCacheableTaskAction(t, true, true)
	tCtx := newTaskExecutionContext(t, taskAction, dataStore)
	require.NoError(t, tCtx.OutputWriter().Put(ctx, ioutils.NewInMemoryOutputReader(&corepb.LiteralMap{}, nil, nil)))

	putCalled := false
	released := false
	r := &TaskActionReconciler{
		DataStore: dataStore,
		Catalog: &stubCatalogClient{
			getFunc: func(context.Context, catalog.Key) (catalog.Entry, error) {
				return catalog.Entry{}, nil
			},
			putFunc: func(_ context.Context, key catalog.Key, reader io.OutputReader, metadata catalog.Metadata) (catalog.Status, error) {
				putCalled = true
				assert.Empty(t, key.CacheKey)
				assert.Equal(t, taskAction.Spec.CacheKey, key.CacheVersion)
				assert.Equal(t, "task-name", metadata.TaskExecutionIdentifier.GetTaskId().GetName())
				return catalog.NewStatus(corepb.CatalogCacheStatus_CACHE_POPULATED, nil), nil
			},
			releaseReservationFunc: func(_ context.Context, _ catalog.Key, ownerID string) error {
				released = true
				assert.Equal(t, "default/cacheable-action", ownerID)
				return nil
			},
		},
	}

	transition := pluginsCore.DoTransition(pluginsCore.PhaseInfoSuccess(&pluginsCore.TaskInfo{}))
	transition, err := r.finalizeCacheAfterExecution(ctx, taskAction, tCtx, transition, false)
	require.NoError(t, err)
	assert.True(t, putCalled)
	assert.True(t, released)
	require.Len(t, transition.Info().Info().ExternalResources, 1)
	assert.Equal(t, corepb.CatalogCacheStatus_CACHE_POPULATED, transition.Info().Info().ExternalResources[0].CacheStatus)
}

func TestHandleCacheAfterExecutionReleasesReservationOnFailure(t *testing.T) {
	ensureTestMetricKeys()
	ctx := context.Background()
	taskAction, dataStore := newCacheableTaskAction(t, true, true)
	tCtx := newTaskExecutionContext(t, taskAction, dataStore)

	released := false
	r := &TaskActionReconciler{
		DataStore: dataStore,
		Catalog: &stubCatalogClient{
			releaseReservationFunc: func(_ context.Context, _ catalog.Key, ownerID string) error {
				released = true
				assert.Equal(t, "default/cacheable-action", ownerID)
				return nil
			},
		},
	}

	transition := pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure("User", "boom", &pluginsCore.TaskInfo{}))
	_, err := r.finalizeCacheAfterExecution(ctx, taskAction, tCtx, transition, false)
	require.NoError(t, err)
	assert.True(t, released)
}

func newCacheableTaskAction(t *testing.T, discoverable bool, serializable bool) (*flyteorgv1.TaskAction, *storage.DataStore) {
	t.Helper()

	dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	require.NoError(t, err)

	taskTemplate := &corepb.TaskTemplate{
		Id: &corepb.Identifier{
			ResourceType: corepb.ResourceType_TASK,
			Project:      "project",
			Domain:       "domain",
			Name:         "task-name",
			Version:      "task-version",
		},
		Type: "python-task",
		Metadata: &corepb.TaskMetadata{
			Discoverable:         discoverable,
			DiscoveryVersion:     "discovery-v1",
			CacheSerializable:    serializable,
			CacheIgnoreInputVars: []string{"ignored_input"},
			Runtime: &corepb.RuntimeMetadata{
				Type: corepb.RuntimeMetadata_FLYTE_SDK,
			},
		},
		Interface: &corepb.TypedInterface{},
		Target: &corepb.TaskTemplate_Container{
			Container: &corepb.Container{
				Image:   "python:3.11",
				Command: []string{"echo"},
			},
		},
	}
	taskTemplateBytes, err := proto.Marshal(taskTemplate)
	require.NoError(t, err)

	taskAction := &flyteorgv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cacheable-action",
			Namespace: "default",
		},
		Spec: flyteorgv1.TaskActionSpec{
			RunName:       "run-name",
			Project:       "project",
			Domain:        "domain",
			ActionName:    "action-name",
			InputURI:      "s3://bucket/inputs.pb",
			RunOutputBase: "s3://bucket/run-output",
			CacheKey:      "precomputed-cache-key",
			TaskType:      "python-task",
			TaskTemplate:  taskTemplateBytes,
		},
	}
	return taskAction, dataStore
}

func newTaskExecutionContext(t *testing.T, taskAction *flyteorgv1.TaskAction, dataStore *storage.DataStore) pluginsCore.TaskExecutionContext {
	t.Helper()

	err := dataStore.WriteProtobuf(context.Background(), storage.DataReference(taskAction.Spec.InputURI), storage.Options{}, &corepb.LiteralMap{})
	require.NoError(t, err)

	tCtx, err := executorplugin.NewTaskExecutionContext(
		taskAction,
		dataStore,
		executorplugin.NewPluginStateManager(nil, 0),
		nil,
		nil,
		nil,
	)
	require.NoError(t, err)
	return tCtx
}
