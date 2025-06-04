package workflowstore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/client/clientset/versioned/fake"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const (
	terminatedTrackingNamespace = "test-ns"
)

func TestTerminatedTrackingStore_Update(t *testing.T) {
	ctx := context.TODO()

	mockClient := fake.NewSimpleClientset().FlyteworkflowV1alpha1()

	scope := promutils.NewTestScope()
	l := &mockWFNamespaceLister{}
	passthroughWfStore := NewPassthroughWorkflowStore(ctx, scope, mockClient, &mockWFLister{V: l})
	wfStore, err := NewTerminatedTrackingStore(ctx, scope, passthroughWfStore)
	assert.NoError(t, err)

	t.Run("Succeeding", func(t *testing.T) {
		name := "succeeding"

		wf := dummyWf(terminatedTrackingNamespace, name)
		wf.Status.Phase = v1alpha1.WorkflowPhaseSucceeding

		_, err := mockClient.FlyteWorkflows(wf.GetNamespace()).Create(ctx, wf, v1.CreateOptions{})
		assert.NoError(t, err)

		_, err = wfStore.Update(ctx, wf)
		assert.NoError(t, err)

		l.GetCb = func(name string) (*v1alpha1.FlyteWorkflow, error) {
			return wf, nil
		}
		succeedingWf, err := wfStore.Get(ctx, terminatedTrackingNamespace, name)
		assert.NoError(t, err)
		assert.Equal(t, succeedingWf, wf)
	})

	t.Run("Terminated", func(t *testing.T) {
		name := "terminated"

		wf := dummyWf(terminatedTrackingNamespace, name)
		wf.Status.Phase = v1alpha1.WorkflowPhaseAborted

		_, err := mockClient.FlyteWorkflows(wf.GetNamespace()).Create(ctx, wf, v1.CreateOptions{})
		assert.NoError(t, err)

		_, err = wfStore.Update(ctx, wf)
		assert.NoError(t, err)

		l.GetCb = func(name string) (*v1alpha1.FlyteWorkflow, error) {
			return wf, nil
		}
		terminatedWf, err := wfStore.Get(ctx, terminatedTrackingNamespace, name)
		assert.Nil(t, terminatedWf)
		assert.True(t, IsWorkflowTerminated(err))
	})
}
