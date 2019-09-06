package workflowstore

import (
	"context"
	"fmt"
	"testing"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/client/clientset/versioned/fake"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
)

func TestResourceVersionCaching_Get_NotInCache(t *testing.T) {
	ctx := context.TODO()
	mockClient := fake.NewSimpleClientset().FlyteworkflowV1alpha1()

	scope := promutils.NewTestScope()
	l := &mockWFNamespaceLister{}
	wfStore := NewResourceVersionCachingStore(ctx, scope, NewPassthroughWorkflowStore(ctx, scope, mockClient, &mockWFLister{V: l}))

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

func TestResourceVersionCaching_Get_UpdateAndRead(t *testing.T) {
	ctx := context.TODO()

	mockClient := fake.NewSimpleClientset().FlyteworkflowV1alpha1()

	namespace := "ns"
	name := "name"
	resourceVersion := "r1"

	wf := dummyWf(namespace, name)
	wf.ResourceVersion = resourceVersion

	t.Run("Stale", func(t *testing.T) {

		scope := promutils.NewTestScope()
		l := &mockWFNamespaceLister{}
		wfStore := NewResourceVersionCachingStore(ctx, scope, NewPassthroughWorkflowStore(ctx, scope, mockClient, &mockWFLister{V: l}))
		// Insert a new workflow with R1
		err := wfStore.Update(ctx, wf, PriorityClassCritical)
		assert.NoError(t, err)

		// Return the same workflow
		l.GetCb = func(name string) (*v1alpha1.FlyteWorkflow, error) {

			return wf, nil
		}

		w, err := wfStore.Get(ctx, namespace, name)
		assert.Error(t, err)
		assert.False(t, IsNotFound(err))
		assert.True(t, IsWorkflowStale(err))
		assert.Nil(t, w)
	})

	t.Run("Updated", func(t *testing.T) {
		scope := promutils.NewTestScope()
		l := &mockWFNamespaceLister{}
		wfStore := NewResourceVersionCachingStore(ctx, scope, NewPassthroughWorkflowStore(ctx, scope, mockClient, &mockWFLister{V: l}))
		// Insert a new workflow with R1
		err := wfStore.Update(ctx, wf, PriorityClassCritical)
		assert.NoError(t, err)

		// Update the workflow version
		wf2 := wf.DeepCopy()
		wf2.ResourceVersion = "r2"

		// Return updated workflow
		l.GetCb = func(name string) (*v1alpha1.FlyteWorkflow, error) {
			return wf2, nil
		}

		w, err := wfStore.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.NotNil(t, w)
		assert.Equal(t, "r2", w.ResourceVersion)
	})
}

func TestResourceVersionCaching_UpdateTerminated(t *testing.T) {
	ctx := context.TODO()

	mockClient := fake.NewSimpleClientset().FlyteworkflowV1alpha1()

	namespace := "ns"
	name := "name"
	resourceVersion := "r1"

	wf := dummyWf(namespace, name)
	wf.ResourceVersion = resourceVersion

	scope := promutils.NewTestScope()
	l := &mockWFNamespaceLister{}
	wfStore := NewResourceVersionCachingStore(ctx, scope, NewPassthroughWorkflowStore(ctx, scope, mockClient, &mockWFLister{V: l}))
	// Insert a new workflow with R1
	err := wfStore.Update(ctx, wf, PriorityClassCritical)
	assert.NoError(t, err)

	rvStore := wfStore.(*resourceVersionCaching)
	v, ok := rvStore.lastUpdatedResourceVersionCache.Load(resourceVersionKey(namespace, name))
	assert.True(t, ok)
	assert.Equal(t, resourceVersion, v.(string))

	wf2 := wf.DeepCopy()
	wf2.Status.Phase = v1alpha1.WorkflowPhaseAborted
	err = wfStore.Update(ctx, wf2, PriorityClassCritical)
	assert.NoError(t, err)

	v, ok = rvStore.lastUpdatedResourceVersionCache.Load(resourceVersionKey(namespace, name))
	assert.False(t, ok)
	assert.Nil(t, v)

}
