package workflowstore

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	listers "github.com/flyteorg/flytepropeller/pkg/client/listers/flyteworkflow/v1alpha1"

	"github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned/fake"

	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
)

type mockWFNamespaceLister struct {
	listers.FlyteWorkflowNamespaceLister
	GetCb func(name string) (*v1alpha1.FlyteWorkflow, error)
}

func (m *mockWFNamespaceLister) Get(name string) (*v1alpha1.FlyteWorkflow, error) {
	return m.GetCb(name)
}

type mockWFLister struct {
	listers.FlyteWorkflowLister
	V listers.FlyteWorkflowNamespaceLister
}

func (m *mockWFLister) FlyteWorkflows(namespace string) listers.FlyteWorkflowNamespaceLister {
	return m.V
}

func TestPassthroughWorkflowStore_Get(t *testing.T) {
	ctx := context.TODO()

	mockClient := fake.NewSimpleClientset().FlyteworkflowV1alpha1()

	l := &mockWFNamespaceLister{}
	wfStore := NewPassthroughWorkflowStore(ctx, promutils.NewTestScope(), mockClient, &mockWFLister{V: l})

	t.Run("notFound", func(t *testing.T) {
		l.GetCb = func(name string) (*v1alpha1.FlyteWorkflow, error) {
			return nil, kubeerrors.NewNotFound(v1alpha1.Resource(v1alpha1.FlyteWorkflowKind), "name")
		}
		w, err := wfStore.Get(ctx, "ns", "name")
		assert.Error(t, err)
		assert.True(t, IsNotFound(err))
		assert.Nil(t, w)
	})

	t.Run("alreadyExists?", func(t *testing.T) {
		l.GetCb = func(name string) (*v1alpha1.FlyteWorkflow, error) {
			return nil, kubeerrors.NewAlreadyExists(v1alpha1.Resource(v1alpha1.FlyteWorkflowKind), "name")
		}
		w, err := wfStore.Get(ctx, "ns", "name")
		assert.Error(t, err)
		assert.Nil(t, w)
	})

	t.Run("unknownError", func(t *testing.T) {
		l.GetCb = func(name string) (*v1alpha1.FlyteWorkflow, error) {
			return nil, fmt.Errorf("error")
		}
		w, err := wfStore.Get(ctx, "ns", "name")
		assert.Error(t, err)
		assert.Nil(t, w)
	})

	t.Run("success", func(t *testing.T) {
		expW := &v1alpha1.FlyteWorkflow{}
		l.GetCb = func(name string) (*v1alpha1.FlyteWorkflow, error) {
			return expW, nil
		}
		w, err := wfStore.Get(ctx, "ns", "name")
		assert.NoError(t, err)
		assert.Equal(t, expW, w)
	})
}

func dummyWf(namespace, name string) *v1alpha1.FlyteWorkflow {
	return &v1alpha1.FlyteWorkflow{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func TestPassthroughWorkflowStore_UpdateStatus(t *testing.T) {

	ctx := context.TODO()

	mockClient := fake.NewSimpleClientset().FlyteworkflowV1alpha1()
	l := &mockWFNamespaceLister{}
	wfStore := NewPassthroughWorkflowStore(ctx, promutils.NewTestScope(), mockClient, &mockWFLister{V: l})

	const namespace = "test-ns"
	t.Run("notFound", func(t *testing.T) {
		wf := dummyWf(namespace, "x")
		_, err := wfStore.UpdateStatus(ctx, wf, PriorityClassCritical)
		assert.NoError(t, err)
		updated, err := mockClient.FlyteWorkflows(namespace).Get(ctx, "x", v1.GetOptions{})
		assert.Error(t, err)
		assert.Nil(t, updated)
	})

	t.Run("Found-Updated", func(t *testing.T) {
		n := mockClient.FlyteWorkflows(namespace)
		wf := dummyWf(namespace, "x")
		wf.GetExecutionStatus().UpdatePhase(v1alpha1.WorkflowPhaseSucceeding, "", nil)
		wf.ResourceVersion = "r1"
		_, err := n.Create(ctx, wf, v1.CreateOptions{})
		assert.NoError(t, err)
		updated, err := n.Get(ctx, "x", v1.GetOptions{})
		if assert.NoError(t, err) {
			assert.Equal(t, v1alpha1.WorkflowPhaseSucceeding, updated.GetExecutionStatus().GetPhase())
			wf.GetExecutionStatus().UpdatePhase(v1alpha1.WorkflowPhaseFailed, "", &core.ExecutionError{})
			_, err := wfStore.UpdateStatus(ctx, wf, PriorityClassCritical)
			assert.NoError(t, err)
			newVal, err := n.Get(ctx, "x", v1.GetOptions{})
			assert.NoError(t, err)
			assert.Equal(t, v1alpha1.WorkflowPhaseFailed, newVal.GetExecutionStatus().GetPhase())
		}
	})

}
