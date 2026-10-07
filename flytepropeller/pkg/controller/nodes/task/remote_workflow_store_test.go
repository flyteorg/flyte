package task

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-test/deep"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func createInmemoryStore(t testing.TB) *storage.DataStore {
	cfg := storage.Config{
		Type: storage.TypeMemory,
	}

	d, err := storage.NewDataStore(&cfg, promutils.NewTestScope())
	assert.NoError(t, err)

	return d
}

func Test_cacheFlyteWorkflow(t *testing.T) {
	store := createInmemoryStore(t)
	t.Run("cache CRD", func(t *testing.T) {
		expected := &v1alpha1.FlyteWorkflow{
			TypeMeta:   v1.TypeMeta{},
			ObjectMeta: v1.ObjectMeta{},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "abc",
				Connections: v1alpha1.Connections{
					Downstream: map[v1alpha1.NodeID][]v1alpha1.NodeID{},
					Upstream:   map[v1alpha1.NodeID][]v1alpha1.NodeID{},
				},
				DeprecatedConnections: v1alpha1.DeprecatedConnections{
					DownstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{},
					UpstreamEdges:   map[v1alpha1.NodeID][]v1alpha1.NodeID{},
				},
			},
		}

		ctx := context.TODO()
		location := storage.DataReference("somekey/file.json")
		r := RemoteFileWorkflowStore{store: store}
		assert.NoError(t, r.PutFlyteWorkflowCRD(ctx, expected, location))
		actual, err := r.GetWorkflowCRD(ctx, location)
		assert.NoError(t, err)
		if diff := deep.Equal(expected, actual); len(diff) > 0 {
			t.Errorf("Cached() Diff = %v\r\n got = %v\r\n want = %v", diff, actual, expected)
		}
	})

	t.Run("cache CRD with deprecatedConnections", func(t *testing.T) {
		expected := &v1alpha1.FlyteWorkflow{
			TypeMeta:   v1.TypeMeta{},
			ObjectMeta: v1.ObjectMeta{},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "abc",
				DeprecatedConnections: v1alpha1.DeprecatedConnections{
					DownstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{},
					UpstreamEdges:   map[v1alpha1.NodeID][]v1alpha1.NodeID{},
				},
			},
			ExecutionConfig: v1alpha1.ExecutionConfig{
				TaskResources: v1alpha1.TaskResources{
					Requests: v1alpha1.TaskResourceSpec{
						CPU:              resource.MustParse("1"),
						Memory:           resource.MustParse("1"),
						Storage:          resource.MustParse("1"),
						EphemeralStorage: resource.MustParse("1"),
						GPU:              resource.MustParse("1"),
					},
					Limits: v1alpha1.TaskResourceSpec{
						CPU:              resource.MustParse("1"),
						Memory:           resource.MustParse("1"),
						Storage:          resource.MustParse("1"),
						EphemeralStorage: resource.MustParse("1"),
						GPU:              resource.MustParse("1"),
					},
				},
			},
		}

		ctx := context.TODO()
		location := storage.DataReference("somekey/file.json")
		r := RemoteFileWorkflowStore{store: store}
		assert.NoError(t, r.PutFlyteWorkflowCRD(ctx, expected, location))
		actual, err := r.GetWorkflowCRD(ctx, location)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("cache compiled workflow", func(t *testing.T) {
		expected := &core.CompiledWorkflowClosure{
			Primary: &core.CompiledWorkflow{
				Template: &core.WorkflowTemplate{
					Id: &core.Identifier{
						ResourceType: core.ResourceType_WORKFLOW,
						Project:      "proj",
						Domain:       "domain",
						Name:         "name",
						Version:      "version",
					},
				},
			},
		}

		ctx := context.TODO()
		location := storage.DataReference("somekey/dynamic_compiled.pb")
		r := RemoteFileWorkflowStore{store: store}
		assert.NoError(t, r.PutCompiledFlyteWorkflow(ctx, expected, location))
		actual, err := r.GetCompiledWorkflow(ctx, location)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(expected, actual))
	})
}

func Test_cacheFlyteWorkflowLargerThanCacheEntry(t *testing.T) {
	// freecache rejects entries larger than 1/1024 of its size, so a 1MB cache can't hold a ~1KB+ CRD.
	store, err := storage.NewDataStore(&storage.Config{
		Type:  storage.TypeMemory,
		Cache: storage.CachingConfig{MaxSizeMegabytes: 1},
	}, promutils.NewTestScope())
	assert.NoError(t, err)

	nodes := map[v1alpha1.NodeID]*v1alpha1.NodeSpec{}
	for i := 0; i < 100; i++ {
		id := fmt.Sprintf("node-%d", i)
		nodes[id] = &v1alpha1.NodeSpec{ID: id, Name: id}
	}
	expected := &v1alpha1.FlyteWorkflow{
		WorkflowSpec: &v1alpha1.WorkflowSpec{
			ID:    "abc",
			Nodes: nodes,
			Connections: v1alpha1.Connections{
				Downstream: map[v1alpha1.NodeID][]v1alpha1.NodeID{},
				Upstream:   map[v1alpha1.NodeID][]v1alpha1.NodeID{},
			},
			DeprecatedConnections: v1alpha1.DeprecatedConnections{
				DownstreamEdges: map[v1alpha1.NodeID][]v1alpha1.NodeID{},
				UpstreamEdges:   map[v1alpha1.NodeID][]v1alpha1.NodeID{},
			},
		},
	}

	ctx := context.TODO()
	r := RemoteFileWorkflowStore{store: store}

	t.Run("put and get CRD", func(t *testing.T) {
		location := storage.DataReference("somekey/futures_compiled.pb")
		assert.NoError(t, r.PutFlyteWorkflowCRD(ctx, expected, location))
		actual, err := r.GetWorkflowCRD(ctx, location)
		assert.NoError(t, err)
		if diff := deep.Equal(expected, actual); len(diff) > 0 {
			t.Errorf("GetWorkflowCRD() Diff = %v", diff)
		}
	})

	t.Run("get missing CRD", func(t *testing.T) {
		_, err := r.GetWorkflowCRD(ctx, storage.DataReference("somekey/missing.pb"))
		assert.Error(t, err)
	})
}
