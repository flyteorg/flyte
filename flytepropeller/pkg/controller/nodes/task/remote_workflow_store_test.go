package task

import (
	"context"
	"testing"

	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
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
	assert.NoError(t, r.Put(ctx, expected, location))
	actual, err := r.Get(ctx, location)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
