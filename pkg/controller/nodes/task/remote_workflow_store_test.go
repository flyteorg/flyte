package task

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
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
