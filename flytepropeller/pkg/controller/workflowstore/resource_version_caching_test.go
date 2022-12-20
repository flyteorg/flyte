package workflowstore

import (
	"context"
	"fmt"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/meta"

	"k8s.io/apimachinery/pkg/runtime"
	testing2 "k8s.io/client-go/testing"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	goObjectHash "github.com/benlaurie/objecthash/go/objecthash"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned/fake"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	resourceVersionCachingNamespace = "ns"
	resourceVersionCachingName      = "name"
)

func init() {
	labeled.SetMetricKeys(contextutils.WorkflowIDKey)
}

func TestResourceVersionCaching_Get_NotInCache(t *testing.T) {
	ctx := context.TODO()
	mockClient := fake.NewSimpleClientset().FlyteworkflowV1alpha1()

	scope := promutils.NewTestScope()
	l := &mockWFNamespaceLister{}
	passthroughWfStore := NewPassthroughWorkflowStore(ctx, scope, mockClient, &mockWFLister{V: l})
	trackTerminatedWfStore, err := NewTerminatedTrackingStore(ctx, scope, passthroughWfStore)
	wfStore := NewResourceVersionCachingStore(ctx, scope, trackTerminatedWfStore)
	assert.NoError(t, err)

	t.Run("notFound", func(t *testing.T) {
		l.GetCb = func(name string) (*v1alpha1.FlyteWorkflow, error) {
			return nil, kubeerrors.NewNotFound(v1alpha1.Resource(v1alpha1.FlyteWorkflowKind), resourceVersionCachingName)
		}
		w, err := wfStore.Get(ctx, resourceVersionCachingNamespace, resourceVersionCachingName)
		assert.Error(t, err)
		assert.True(t, IsNotFound(err))
		assert.Nil(t, w)
	})

	t.Run("alreadyExists?", func(t *testing.T) {
		l.GetCb = func(name string) (*v1alpha1.FlyteWorkflow, error) {
			return nil, kubeerrors.NewAlreadyExists(v1alpha1.Resource(v1alpha1.FlyteWorkflowKind), resourceVersionCachingName)
		}
		w, err := wfStore.Get(ctx, resourceVersionCachingNamespace, resourceVersionCachingName)
		assert.Error(t, err)
		assert.Nil(t, w)
	})

	t.Run("unknownError", func(t *testing.T) {
		l.GetCb = func(name string) (*v1alpha1.FlyteWorkflow, error) {
			return nil, fmt.Errorf("error")
		}
		w, err := wfStore.Get(ctx, resourceVersionCachingNamespace, resourceVersionCachingName)
		assert.Error(t, err)
		assert.Nil(t, w)
	})

	t.Run("success", func(t *testing.T) {
		expW := &v1alpha1.FlyteWorkflow{}
		l.GetCb = func(name string) (*v1alpha1.FlyteWorkflow, error) {
			return expW, nil
		}
		w, err := wfStore.Get(ctx, resourceVersionCachingNamespace, resourceVersionCachingName)
		assert.NoError(t, err)
		assert.Equal(t, expW, w)
	})
}

func fromHashToByteArray(input [32]byte) []byte {
	output := make([]byte, 32)
	for idx, val := range input { //nolint
		output[idx] = val
	}
	return output
}

func objHash(obj interface{}) string {
	jsonObj, err := goObjectHash.CommonJSONify(obj)
	if err != nil {
		panic(err)
	}

	raw, err := goObjectHash.ObjectHash(jsonObj)
	if err != nil {
		panic(err)
	}

	return string(fromHashToByteArray(raw))
}

func createFakeClientSet() *fake.Clientset {
	fakeClientSet := fake.NewSimpleClientset()
	// Override Update action to modify the resource version to simulate real updates.
	fakeClientSet.PrependReactor("update", "*",
		func(action testing2.Action) (handled bool, ret runtime.Object, err error) {
			switch action := action.(type) {
			case testing2.UpdateAction:
				newObj := action.GetObject().DeepCopyObject()
				a, err := meta.Accessor(newObj)
				if err != nil {
					return false, nil, err
				}

				origObj, err := fakeClientSet.Tracker().Get(action.GetResource(), a.GetNamespace(), a.GetName())
				if err != nil {
					return false, nil, err
				}

				if origObj == nil {
					return false, newObj, nil
				}

				origHash := objHash(origObj)
				newHash := objHash(newObj)

				if origHash != newHash {
					a.SetResourceVersion(a.GetResourceVersion() + ".new")
				}

				err = fakeClientSet.Tracker().Update(action.GetResource(), newObj, a.GetNamespace())
				return true, newObj, err
			}

			return false, ret, nil
		})

	return fakeClientSet
}

func TestResourceVersionCaching_Get_UpdateAndRead(t *testing.T) {
	ctx := context.TODO()

	resourceVersion := "r1"

	mockClient := createFakeClientSet().FlyteworkflowV1alpha1()

	t.Run("Stale", func(t *testing.T) {
		staleName := resourceVersionCachingName + ".stale"
		wf := dummyWf(resourceVersionCachingNamespace, staleName)
		wf.ResourceVersion = resourceVersion

		scope := promutils.NewTestScope()
		l := &mockWFNamespaceLister{}
		// Return the same workflow
		l.GetCb = func(name string) (*v1alpha1.FlyteWorkflow, error) {
			return wf, nil
		}

		_, err := mockClient.FlyteWorkflows(wf.GetNamespace()).Create(ctx, wf, v1.CreateOptions{})
		assert.NoError(t, err)

		newWf := wf.DeepCopy()
		newWf.Status.Phase = v1alpha1.WorkflowPhaseSucceeding

		passthroughWfStore := NewPassthroughWorkflowStore(ctx, scope, mockClient, &mockWFLister{V: l})
		trackTerminatedWfStore, err := NewTerminatedTrackingStore(ctx, scope, passthroughWfStore)
		wfStore := NewResourceVersionCachingStore(ctx, scope, trackTerminatedWfStore)
		assert.NoError(t, err)
		// Insert a new workflow with R1
		_, err = wfStore.Update(ctx, newWf, PriorityClassCritical)
		assert.NoError(t, err)

		w, err := wfStore.Get(ctx, resourceVersionCachingNamespace, staleName)
		assert.Error(t, err)
		assert.False(t, IsNotFound(err))
		assert.True(t, IsWorkflowStale(err))
		assert.Nil(t, w)
	})

	t.Run("Updated", func(t *testing.T) {
		updatedName := resourceVersionCachingName + ".updated"
		wf := dummyWf(resourceVersionCachingNamespace, updatedName)
		wf.ResourceVersion = resourceVersion

		scope := promutils.NewTestScope()
		l := &mockWFNamespaceLister{}
		passthroughWfStore := NewPassthroughWorkflowStore(ctx, scope, mockClient, &mockWFLister{V: l})
		trackTerminatedWfStore, err := NewTerminatedTrackingStore(ctx, scope, passthroughWfStore)
		wfStore := NewResourceVersionCachingStore(ctx, scope, trackTerminatedWfStore)
		assert.NoError(t, err)
		// Insert a new workflow with R1
		_, err = wfStore.Update(ctx, wf, PriorityClassCritical)
		assert.NoError(t, err)

		// Update the workflow version
		wf2 := wf.DeepCopy()
		wf2.ResourceVersion = "r2"

		// Return updated workflow
		l.GetCb = func(name string) (*v1alpha1.FlyteWorkflow, error) {
			return wf2, nil
		}

		w, err := wfStore.Get(ctx, resourceVersionCachingNamespace, updatedName)
		assert.NoError(t, err)
		assert.NotNil(t, w)
		assert.Equal(t, "r2", w.ResourceVersion)
	})

	// If we -mistakenly- attempted to update the store with the exact workflow object, etcd. will not bump the
	// resource version. Next read operation should continue to retrieve the same instance of the object.
	t.Run("NotUpdated", func(t *testing.T) {
		notUpdatedName := resourceVersionCachingName + ".not-updated"
		wf := dummyWf(resourceVersionCachingNamespace, notUpdatedName)
		wf.ResourceVersion = resourceVersion

		_, err := mockClient.FlyteWorkflows(wf.GetNamespace()).Create(ctx, wf, v1.CreateOptions{})
		assert.NoError(t, err)

		scope := promutils.NewTestScope()
		l := &mockWFNamespaceLister{}
		passthroughWfStore := NewPassthroughWorkflowStore(ctx, scope, mockClient, &mockWFLister{V: l})
		trackTerminatedWfStore, err := NewTerminatedTrackingStore(ctx, scope, passthroughWfStore)
		wfStore := NewResourceVersionCachingStore(ctx, scope, trackTerminatedWfStore)
		assert.NoError(t, err)
		// Insert a new workflow with R1
		_, err = wfStore.Update(ctx, wf, PriorityClassCritical)
		assert.NoError(t, err)

		// Return updated workflow
		l.GetCb = func(name string) (*v1alpha1.FlyteWorkflow, error) {
			return wf, nil
		}

		w, err := wfStore.Get(ctx, resourceVersionCachingNamespace, notUpdatedName)
		assert.NoError(t, err)
		assert.NotNil(t, w)
		assert.Equal(t, "r1", w.ResourceVersion)
	})
}

func TestResourceVersionCaching_UpdateTerminated(t *testing.T) {
	ctx := context.TODO()

	mockClient := createFakeClientSet().FlyteworkflowV1alpha1()

	resourceVersion := "r1"

	wf := dummyWf(resourceVersionCachingNamespace, resourceVersionCachingName)
	wf.ResourceVersion = resourceVersion

	_, err := mockClient.FlyteWorkflows(wf.GetNamespace()).Create(ctx, wf, v1.CreateOptions{})
	assert.NoError(t, err)

	newWf := wf.DeepCopy()
	newWf.Status.Phase = v1alpha1.WorkflowPhaseSucceeding

	scope := promutils.NewTestScope()
	l := &mockWFNamespaceLister{}
	passthroughWfStore := NewPassthroughWorkflowStore(ctx, scope, mockClient, &mockWFLister{V: l})
	trackTerminatedWfStore, err := NewTerminatedTrackingStore(ctx, scope, passthroughWfStore)
	wfStore := NewResourceVersionCachingStore(ctx, scope, trackTerminatedWfStore)
	assert.NoError(t, err)
	// Insert a new workflow with R1
	_, err = wfStore.Update(ctx, newWf, PriorityClassCritical)
	assert.NoError(t, err)

	rvStore := wfStore.(*resourceVersionCaching)
	v, ok := rvStore.lastUpdatedResourceVersionCache.Load(resourceVersionKey(resourceVersionCachingNamespace, resourceVersionCachingName))
	assert.True(t, ok)
	assert.Equal(t, resourceVersion, v.(string))

	wf2 := wf.DeepCopy()
	wf2.Status.Phase = v1alpha1.WorkflowPhaseAborted
	_, err = wfStore.Update(ctx, wf2, PriorityClassCritical)
	assert.NoError(t, err)

	v, ok = rvStore.lastUpdatedResourceVersionCache.Load(resourceVersionKey(resourceVersionCachingNamespace, resourceVersionCachingName))
	assert.False(t, ok)
	assert.Nil(t, v)

	// validate that terminated workflows are not retrievable
	terminatedWf, err := wfStore.Get(ctx, resourceVersionCachingNamespace, resourceVersionCachingName)
	assert.Nil(t, terminatedWf)
	assert.True(t, IsWorkflowTerminated(err))
}
