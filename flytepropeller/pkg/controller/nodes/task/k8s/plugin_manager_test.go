package k8s

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"k8s.io/client-go/kubernetes/scheme"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/backoff"
	"github.com/flyteorg/flytestdlib/promutils"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	pluginsCoreMock "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	pluginsk8sMock "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"

	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/flyteorg/flytepropeller/pkg/controller/executors/mocks"
)

type extendedFakeClient struct {
	client.Client
	CreateError error
	GetError    error
	DeleteError error
	PatchError  error
	UpdateError error
}

func (e extendedFakeClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if e.CreateError != nil {
		return e.CreateError
	}
	return e.Client.Create(ctx, obj)
}

func (e extendedFakeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	if e.GetError != nil {
		return e.GetError
	}
	return e.Client.Get(ctx, key, obj)
}

func (e extendedFakeClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	if e.DeleteError != nil {
		return e.DeleteError
	}

	return e.Client.Delete(ctx, obj, opts...)
}

func (e extendedFakeClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	if e.PatchError != nil {
		return e.PatchError
	}

	return e.Client.Patch(ctx, obj, patch, opts...)
}

func (e extendedFakeClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if e.UpdateError != nil {
		return e.UpdateError
	}

	return e.Client.Update(ctx, obj, opts...)
}

type k8sSampleHandler struct {
}

func (k8sSampleHandler) GetProperties() k8s.PluginProperties {
	panic("implement me")
}

func (k8sSampleHandler) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {
	panic("implement me")
}

func (k8sSampleHandler) BuildIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	panic("implement me")
}

func (k8sSampleHandler) GetTaskPhase(ctx context.Context, pluginContext k8s.PluginContext, resource client.Object) (pluginsCore.PhaseInfo, error) {
	panic("implement me")
}

type pluginWithAbortOverride struct {
	mock.Mock
}

func (p *pluginWithAbortOverride) GetProperties() k8s.PluginProperties {
	return p.Called().Get(0).(k8s.PluginProperties)
}

func (p *pluginWithAbortOverride) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (client.Object, error) {
	panic("implement me")
}

func (p *pluginWithAbortOverride) BuildIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	args := p.Called(ctx, taskCtx)
	return args.Get(0).(client.Object), args.Error(1)
}

func (p *pluginWithAbortOverride) GetTaskPhase(ctx context.Context, pluginContext k8s.PluginContext, resource client.Object) (pluginsCore.PhaseInfo, error) {
	panic("implement me")
}

func (p *pluginWithAbortOverride) OnAbort(ctx context.Context, tCtx pluginsCore.TaskExecutionContext, resource client.Object) (behavior k8s.AbortBehavior, err error) {
	args := p.Called(ctx, tCtx, resource)
	return args.Get(0).(k8s.AbortBehavior), args.Error(1)
}

func ExampleNewPluginManager() {
	sCtx := &pluginsCoreMock.SetupContext{}
	fakeKubeClient := mocks.NewFakeKubeClient()
	sCtx.On("KubeClient").Return(fakeKubeClient)
	sCtx.On("OwnerKind").Return("test")
	sCtx.On("EnqueueOwner").Return(pluginsCore.EnqueueOwner(func(name k8stypes.NamespacedName) error { return nil }))
	sCtx.On("MetricsScope").Return(promutils.NewTestScope())
	ctx := context.TODO()
	exec, err := NewPluginManager(ctx, sCtx, k8s.PluginEntry{
		ID:                  "SampleHandler",
		RegisteredTaskTypes: []pluginsCore.TaskType{"container"},
		ResourceToWatch:     &v1.Pod{},
		Plugin:              k8sSampleHandler{},
	}, NewResourceMonitorIndex())
	if err == nil {
		fmt.Printf("Created executor: %v\n", exec.GetID())
	} else {
		fmt.Printf("Error in creating executor: %s\n", err.Error())
	}

	// Output:
	// Created executor: SampleHandler
}

type dummyOutputWriter struct {
	io.OutputWriter
	r io.OutputReader
}

func (d *dummyOutputWriter) Put(ctx context.Context, reader io.OutputReader) error {
	d.r = reader
	return nil
}

func getMockTaskContext(initPhase PluginPhase, wantPhase PluginPhase) pluginsCore.TaskExecutionContext {
	taskExecutionContext := &pluginsCoreMock.TaskExecutionContext{}
	taskExecutionContext.OnTaskExecutionMetadata().Return(getMockTaskExecutionMetadata())

	tReader := &pluginsCoreMock.TaskReader{}
	tReader.OnReadMatch(mock.Anything).Return(&core.TaskTemplate{}, nil)
	taskExecutionContext.OnTaskReader().Return(tReader)

	customStateReader := &pluginsCoreMock.PluginStateReader{}
	customStateReader.OnGetMatch(mock.MatchedBy(func(i interface{}) bool {
		ps, ok := i.(*PluginState)
		if ok {
			ps.Phase = initPhase
			return true
		}
		return false
	})).Return(uint8(0), nil)
	taskExecutionContext.OnPluginStateReader().Return(customStateReader)

	customStateWriter := &pluginsCoreMock.PluginStateWriter{}
	customStateWriter.OnPutMatch(mock.Anything, mock.MatchedBy(func(i interface{}) bool {
		ps, ok := i.(*PluginState)
		return ok && ps.Phase == wantPhase
	})).Return(nil)
	taskExecutionContext.OnPluginStateWriter().Return(customStateWriter)
	taskExecutionContext.OnOutputWriter().Return(&dummyOutputWriter{})

	taskExecutionContext.OnDataStore().Return(nil)
	taskExecutionContext.OnMaxDatasetSizeBytes().Return(int64(0))
	return taskExecutionContext
}

func getMockTaskExecutionMetadata() pluginsCore.TaskExecutionMetadata {
	taskExecutionMetadata := &pluginsCoreMock.TaskExecutionMetadata{}
	taskExecutionMetadata.On("GetNamespace").Return("ns")
	taskExecutionMetadata.On("GetAnnotations").Return(map[string]string{"aKey": "aVal"})
	taskExecutionMetadata.On("GetLabels").Return(map[string]string{"lKey": "lVal"})
	taskExecutionMetadata.On("GetOwnerReference").Return(v12.OwnerReference{Name: "x"})

	id := &pluginsCoreMock.TaskExecutionID{}
	id.On("GetGeneratedName").Return("test")
	id.On("GetID").Return(core.TaskExecutionIdentifier{})
	taskExecutionMetadata.On("GetTaskExecutionID").Return(id)
	return taskExecutionMetadata
}

func getMockTaskExecutionMetadataCustom(
	tid string,
	ns string,
	annotations map[string]string,
	labels map[string]string,
	ownerRef v12.OwnerReference) pluginsCore.TaskExecutionMetadata {
	taskExecutionMetadata := &pluginsCoreMock.TaskExecutionMetadata{}
	taskExecutionMetadata.On("GetNamespace").Return(ns)
	taskExecutionMetadata.On("GetAnnotations").Return(annotations)
	taskExecutionMetadata.On("GetLabels").Return(labels)
	taskExecutionMetadata.On("GetOwnerReference").Return(ownerRef)

	id := &pluginsCoreMock.TaskExecutionID{}
	id.On("GetGeneratedName").Return(tid)
	id.On("GetID").Return(core.TaskExecutionIdentifier{})
	taskExecutionMetadata.On("GetTaskExecutionID").Return(id)
	return taskExecutionMetadata
}

func dummySetupContext(fakeClient client.Client) pluginsCore.SetupContext {
	setupContext := &pluginsCoreMock.SetupContext{}
	var enqueueOwnerFunc = pluginsCore.EnqueueOwner(func(ownerId k8stypes.NamespacedName) error { return nil })
	setupContext.On("EnqueueOwner").Return(enqueueOwnerFunc)

	kubeClient := &pluginsCoreMock.KubeClient{}
	kubeClient.On("GetClient").Return(fakeClient)
	kubeClient.On("GetCache").Return(&mocks.FakeInformers{})
	setupContext.On("KubeClient").Return(kubeClient)

	setupContext.On("OwnerKind").Return("x")
	setupContext.On("MetricsScope").Return(promutils.NewTestScope())

	return setupContext
}

func buildPluginWithAbortOverride(ctx context.Context, tctx pluginsCore.TaskExecutionContext, abortBehavior k8s.AbortBehavior, client client.Client) (*PluginManager, error) {
	pluginResource := &v1.Pod{}

	mockResourceHandler := new(pluginWithAbortOverride)

	mockResourceHandler.On(
		"OnAbort", ctx, tctx, pluginResource,
	).Return(abortBehavior, nil)

	mockResourceHandler.On(
		"BuildIdentityResource", ctx, tctx.TaskExecutionMetadata(),
	).Return(pluginResource, nil)

	mockResourceHandler.On("GetProperties").Return(k8s.PluginProperties{})

	mockClient := extendedFakeClient{
		Client: client,
	}

	return NewPluginManager(ctx, dummySetupContext(mockClient), k8s.PluginEntry{
		ID:              "x",
		ResourceToWatch: pluginResource,
		Plugin:          mockResourceHandler,
	}, NewResourceMonitorIndex())
}

func TestK8sTaskExecutor_Handle_LaunchResource(t *testing.T) {
	ctx := context.TODO()
	/*var tmpl *core.TaskTemplate
	var inputs *core.LiteralMap*/

	t.Run("jobQueued", func(t *testing.T) {
		tCtx := getMockTaskContext(PluginPhaseNotStarted, PluginPhaseStarted)
		// common setup code
		mockResourceHandler := &pluginsk8sMock.Plugin{}
		mockResourceHandler.OnGetProperties().Return(k8s.PluginProperties{})
		mockResourceHandler.OnBuildResourceMatch(mock.Anything, mock.Anything).Return(&v1.Pod{}, nil)
		fakeClient := fake.NewClientBuilder().WithRuntimeObjects().Build()
		pluginManager, err := NewPluginManager(ctx, dummySetupContext(fakeClient), k8s.PluginEntry{
			ID:              "x",
			ResourceToWatch: &v1.Pod{},
			Plugin:          mockResourceHandler,
		}, NewResourceMonitorIndex())
		assert.NoError(t, err)

		transition, err := pluginManager.Handle(ctx, tCtx)
		assert.NoError(t, err)
		assert.NotNil(t, transition)
		transitionInfo := transition.Info()
		assert.NotNil(t, transitionInfo)
		assert.Equal(t, pluginsCore.PhaseQueued, transitionInfo.Phase())
		createdPod := &v1.Pod{}

		pluginManager.AddObjectMetadata(tCtx.TaskExecutionMetadata(), createdPod, &config.K8sPluginConfig{})
		assert.NoError(t, fakeClient.Get(ctx, k8stypes.NamespacedName{Namespace: tCtx.TaskExecutionMetadata().GetNamespace(),
			Name: tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()}, createdPod))
		assert.Equal(t, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), createdPod.Name)
		assert.NoError(t, fakeClient.Delete(ctx, createdPod))
	})

	t.Run("jobAlreadyExists", func(t *testing.T) {
		tctx := getMockTaskContext(PluginPhaseNotStarted, PluginPhaseStarted)
		// common setup code
		mockResourceHandler := &pluginsk8sMock.Plugin{}
		mockResourceHandler.OnGetProperties().Return(k8s.PluginProperties{})
		mockResourceHandler.OnBuildResourceMatch(mock.Anything, mock.Anything).Return(&v1.Pod{}, nil)
		fakeClient := fake.NewClientBuilder().WithRuntimeObjects().Build()
		pluginManager, err := NewPluginManager(ctx, dummySetupContext(fakeClient), k8s.PluginEntry{
			ID:              "x",
			ResourceToWatch: &v1.Pod{},
			Plugin:          mockResourceHandler,
		}, NewResourceMonitorIndex())
		assert.NoError(t, err)

		createdPod := &v1.Pod{}
		pluginManager.AddObjectMetadata(tctx.TaskExecutionMetadata(), createdPod, &config.K8sPluginConfig{})
		assert.NoError(t, fakeClient.Create(ctx, createdPod))

		transition, err := pluginManager.Handle(ctx, tctx)
		assert.NoError(t, err)
		assert.NotNil(t, transition)
		transitionInfo := transition.Info()
		assert.NotNil(t, transitionInfo)
		assert.Equal(t, pluginsCore.PhaseQueued, transitionInfo.Phase())

		assert.NoError(t, fakeClient.Get(ctx, k8stypes.NamespacedName{Namespace: tctx.TaskExecutionMetadata().GetNamespace(),
			Name: tctx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()}, createdPod))
		assert.Equal(t, tctx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), createdPod.Name)
		assert.NoError(t, fakeClient.Delete(ctx, createdPod))
	})

	t.Run("jobQuotaExceeded", func(t *testing.T) {
		tctx := getMockTaskContext(PluginPhaseNotStarted, PluginPhaseNotStarted)
		// common setup code
		mockResourceHandler := &pluginsk8sMock.Plugin{}
		mockResourceHandler.OnGetProperties().Return(k8s.PluginProperties{})
		mockResourceHandler.OnBuildResourceMatch(mock.Anything, mock.Anything).Return(&v1.Pod{}, nil)
		fakeClient := extendedFakeClient{
			Client:      fake.NewClientBuilder().WithRuntimeObjects().Build(),
			CreateError: k8serrors.NewForbidden(schema.GroupResource{}, "", errors.New("exceeded quota")),
		}

		pluginManager, err := NewPluginManager(ctx, dummySetupContext(fakeClient), k8s.PluginEntry{
			ID:              "x",
			ResourceToWatch: &v1.Pod{},
			Plugin:          mockResourceHandler,
		}, NewResourceMonitorIndex())
		assert.NoError(t, err)

		createdPod := &v1.Pod{}
		transition, err := pluginManager.Handle(ctx, tctx)
		assert.NoError(t, err)
		assert.NotNil(t, transition)
		transitionInfo := transition.Info()
		assert.NotNil(t, transitionInfo)
		assert.Equal(t, pluginsCore.PhaseWaitingForResources, transitionInfo.Phase())

		err = fakeClient.Get(ctx, k8stypes.NamespacedName{Namespace: tctx.TaskExecutionMetadata().GetNamespace(),
			Name: tctx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()}, createdPod)
		assert.Error(t, err)
		t.Log(err)
		assert.True(t, k8serrors.IsNotFound(err))
	})

	t.Run("jobForbidden", func(t *testing.T) {
		// common setup code
		tctx := getMockTaskContext(PluginPhaseNotStarted, PluginPhaseNotStarted)
		// common setup code
		mockResourceHandler := &pluginsk8sMock.Plugin{}
		mockResourceHandler.OnGetProperties().Return(k8s.PluginProperties{})
		mockResourceHandler.OnBuildResourceMatch(mock.Anything, mock.Anything).Return(&v1.Pod{}, nil)
		fakeClient := extendedFakeClient{
			Client:      fake.NewClientBuilder().WithRuntimeObjects().Build(),
			CreateError: k8serrors.NewForbidden(schema.GroupResource{}, "", errors.New("auth error")),
		}

		pluginManager, err := NewPluginManager(ctx, dummySetupContext(fakeClient), k8s.PluginEntry{
			ID:              "x",
			ResourceToWatch: &v1.Pod{},
			Plugin:          mockResourceHandler,
		}, NewResourceMonitorIndex())
		assert.NoError(t, err)

		createdPod := &v1.Pod{}
		transition, err := pluginManager.Handle(ctx, tctx)
		assert.NoError(t, err)
		assert.NotNil(t, transition)
		transitionInfo := transition.Info()
		assert.NotNil(t, transitionInfo)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, transitionInfo.Phase())

		err = fakeClient.Get(ctx, k8stypes.NamespacedName{Namespace: tctx.TaskExecutionMetadata().GetNamespace(),
			Name: tctx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()}, createdPod)
		assert.Error(t, err)
		assert.True(t, k8serrors.IsNotFound(err))
	})

	t.Run("Insufficient resource blocking pod creation for the first time", func(t *testing.T) {
		tctx := getMockTaskContext(PluginPhaseNotStarted, PluginPhaseNotStarted)
		// Creating a mock k8s plugin
		mockResourceHandler := &pluginsk8sMock.Plugin{}
		mockResourceHandler.OnGetProperties().Return(k8s.PluginProperties{})
		mockResourceHandler.OnBuildResourceMatch(mock.Anything, mock.Anything).Return(&v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       flytek8s.PodKind,
				APIVersion: v1.SchemeGroupVersion.String(),
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{v1.ResourceCPU: resource.MustParse("1"), v1.ResourceMemory: resource.MustParse("1Gi")},
					}},
					{Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{v1.ResourceCPU: resource.MustParse("2"), v1.ResourceMemory: resource.MustParse("2Gi")},
					}},
					{Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{v1.ResourceCPU: resource.MustParse("3"), v1.ResourceMemory: resource.MustParse("3Gi")},
					}},
				},
			},
		}, nil)
		fakeClient := extendedFakeClient{
			Client: fake.NewClientBuilder().WithRuntimeObjects().Build(),
			CreateError: k8serrors.NewForbidden(schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: limits.memory=3Gi, "+
				"used: limits.memory=7976Gi, limited: limits.memory=8000Gi")),
		}

		backOffController := backoff.NewController(ctx)
		pluginManager, err := NewPluginManagerWithBackOff(ctx, dummySetupContext(fakeClient), k8s.PluginEntry{
			ID:              "x",
			ResourceToWatch: &v1.Pod{},
			Plugin:          mockResourceHandler,
		}, backOffController, NewResourceMonitorIndex())

		assert.NoError(t, err)
		transition, err := pluginManager.Handle(ctx, tctx)
		assert.NoError(t, err)
		assert.NotNil(t, transition)
		assert.Equal(t, transition.Info().Phase(), pluginsCore.PhaseWaitingForResources)

		// Build a reference resource that is supposed to be identical to the resource built by pluginManager
		referenceResource, err := mockResourceHandler.BuildResource(ctx, tctx)
		assert.NoError(t, err)
		pluginManager.AddObjectMetadata(tctx.TaskExecutionMetadata(), referenceResource, config.GetK8sPluginConfig())
		refKey := backoff.ComposeResourceKey(referenceResource)
		podBackOffHandler, found := backOffController.GetBackOffHandler(refKey)
		assert.True(t, found)
		assert.Equal(t, uint32(1), podBackOffHandler.BackOffExponent.Load())
	})
}

func TestPluginManager_Abort(t *testing.T) {
	ctx := context.TODO()
	tm := getMockTaskExecutionMetadata()
	res := &v1.Pod{
		ObjectMeta: v12.ObjectMeta{
			Name:      tm.GetTaskExecutionID().GetGeneratedName(),
			Namespace: tm.GetNamespace(),
		},
	}

	t.Run("Abort Pod Exists", func(t *testing.T) {
		// common setup code
		tctx := getMockTaskContext(PluginPhaseStarted, PluginPhaseStarted)
		fc := extendedFakeClient{Client: fake.NewFakeClientWithScheme(scheme.Scheme, res)}

		// common setup code
		mockResourceHandler := &pluginsk8sMock.Plugin{}
		mockResourceHandler.OnGetProperties().Return(k8s.PluginProperties{})
		mockResourceHandler.OnBuildIdentityResourceMatch(mock.Anything, tctx.TaskExecutionMetadata()).Return(&v1.Pod{}, nil)
		mockResourceHandler.OnGetTaskPhaseMatch(mock.Anything, mock.Anything, mock.Anything).Return(pluginsCore.PhaseInfo{}, nil)
		pluginManager, err := NewPluginManager(ctx, dummySetupContext(fc), k8s.PluginEntry{
			ID:              "x",
			ResourceToWatch: &v1.Pod{},
			Plugin:          mockResourceHandler,
		}, NewResourceMonitorIndex())
		assert.NotNil(t, res)
		assert.NoError(t, err)

		err = pluginManager.Abort(ctx, tctx)
		assert.NoError(t, err)

		// no custom cleanup policy has been specified
		mockResourceHandler.AssertNumberOfCalls(t, "OnAbort", 0)
	})

	t.Run("Abort Pod doesn't exist", func(t *testing.T) {
		// common setup code
		tctx := getMockTaskContext(PluginPhaseStarted, PluginPhaseStarted)
		fc := extendedFakeClient{Client: fake.NewFakeClientWithScheme(scheme.Scheme)}
		// common setup code
		mockResourceHandler := &pluginsk8sMock.Plugin{}
		mockResourceHandler.OnGetProperties().Return(k8s.PluginProperties{})
		mockResourceHandler.OnBuildIdentityResourceMatch(mock.Anything, tctx.TaskExecutionMetadata()).Return(&v1.Pod{}, nil)
		mockResourceHandler.OnGetTaskPhaseMatch(mock.Anything, mock.Anything, mock.Anything).Return(pluginsCore.PhaseInfo{}, nil)
		pluginManager, err := NewPluginManager(ctx, dummySetupContext(fc), k8s.PluginEntry{
			ID:              "x",
			ResourceToWatch: &v1.Pod{},
			Plugin:          mockResourceHandler,
		}, NewResourceMonitorIndex())
		assert.NotNil(t, res)
		assert.NoError(t, err)

		err = pluginManager.Abort(ctx, tctx)
		assert.NoError(t, err)
	})

	t.Run("Abort Plugin has Patch PluginAbortOverride", func(t *testing.T) {
		tctx := getMockTaskContext(PluginPhaseStarted, PluginPhaseStarted)
		expectedErr := errors.New("client-side patch error")
		pluginManager, err := buildPluginWithAbortOverride(
			ctx,
			tctx,
			k8s.AbortBehaviorPatchDefaultResource(k8s.PatchResourceOperation{
				Patch:   nil,
				Options: nil,
			}, false),
			extendedFakeClient{
				DeleteError: errors.New(
					"kubeClient.Delete() should not be called if custom cleanup policy exists"),
				PatchError: expectedErr,
			})

		assert.NotNil(t, res)
		assert.NoError(t, err)

		err = pluginManager.Abort(ctx, tctx)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("Abort Plugin has Update PluginAbortOverride", func(t *testing.T) {
		tctx := getMockTaskContext(PluginPhaseStarted, PluginPhaseStarted)
		expectedErr := errors.New("client-side update error")
		pluginManager, err := buildPluginWithAbortOverride(
			ctx,
			tctx,
			k8s.AbortBehaviorUpdateDefaultResource(k8s.UpdateResourceOperation{
				Options: nil,
			}, false),
			extendedFakeClient{
				DeleteError: errors.New(
					"kubeClient.Delete() should not be called if custom cleanup policy exists"),
				UpdateError: expectedErr,
			})

		assert.NotNil(t, res)
		assert.NoError(t, err)

		err = pluginManager.Abort(ctx, tctx)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("Abort Plugin has Delete PluginAbortOverride", func(t *testing.T) {
		tctx := getMockTaskContext(PluginPhaseStarted, PluginPhaseStarted)
		expectedErr := errors.New("client-side delete error")
		pluginManager, err := buildPluginWithAbortOverride(
			ctx,
			tctx,
			k8s.AbortBehaviorDeleteDefaultResource(),
			extendedFakeClient{
				DeleteError: expectedErr,
			})

		assert.NotNil(t, res)
		assert.NoError(t, err)

		err = pluginManager.Abort(ctx, tctx)
		assert.Equal(t, expectedErr, err)
	})
}

func TestPluginManager_Handle_CheckResourceStatus(t *testing.T) {
	ctx := context.TODO()
	tm := getMockTaskExecutionMetadata()
	res := &v1.Pod{
		ObjectMeta: v12.ObjectMeta{
			Name:      tm.GetTaskExecutionID().GetGeneratedName(),
			Namespace: tm.GetNamespace(),
		},
	}
	type args struct {
		getTaskPhaseCB func() (pluginsCore.PhaseInfo, error)
		fakeClient     func() extendedFakeClient
	}

	type want struct {
		wantPhase        pluginsCore.Phase
		wantErr          bool
		wantOutputReader bool
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"lookup-not-found",
			args{
				fakeClient: func() extendedFakeClient {
					return extendedFakeClient{GetError: k8serrors.NewNotFound(schema.GroupResource{}, "")}
				},
				getTaskPhaseCB: func() (pluginsCore.PhaseInfo, error) {
					return pluginsCore.PhaseInfo{}, fmt.Errorf("not expected")
				},
			},
			want{
				wantPhase: pluginsCore.PhaseRetryableFailure,
			},
		},
		{
			"lookup-error",
			args{
				fakeClient: func() extendedFakeClient {
					return extendedFakeClient{GetError: k8serrors.NewForbidden(schema.GroupResource{}, "", fmt.Errorf("error"))}
				},
				getTaskPhaseCB: func() (pluginsCore.PhaseInfo, error) {
					return pluginsCore.PhaseInfo{}, fmt.Errorf("not expected")
				},
			},
			want{
				wantPhase: pluginsCore.PhaseUndefined,
				wantErr:   true,
			},
		},
		{
			"lookup-success-complete",
			args{
				fakeClient: func() extendedFakeClient {
					return extendedFakeClient{Client: fake.NewFakeClient(res)}
				},
				getTaskPhaseCB: func() (pluginsCore.PhaseInfo, error) {
					return pluginsCore.PhaseInfoSuccess(nil), nil
				},
			},
			want{
				wantPhase:        pluginsCore.PhaseSuccess,
				wantOutputReader: true,
			},
		},
		{
			"lookup-success-running",
			args{
				fakeClient: func() extendedFakeClient {
					return extendedFakeClient{Client: fake.NewFakeClient(res)}
				},
				getTaskPhaseCB: func() (pluginsCore.PhaseInfo, error) {
					return pluginsCore.PhaseInfoRunning(4, nil), nil
				},
			},
			want{
				wantPhase: pluginsCore.PhaseRunning,
			},
		},
		{
			"lookup-success-error",
			args{
				fakeClient: func() extendedFakeClient {
					return extendedFakeClient{Client: fake.NewFakeClient(res)}
				},
				getTaskPhaseCB: func() (pluginsCore.PhaseInfo, error) {
					return pluginsCore.PhaseInfoUndefined, fmt.Errorf("error")
				},
			},
			want{
				wantPhase: pluginsCore.PhaseUndefined,
				wantErr:   true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// common setup code
			tctx := getMockTaskContext(PluginPhaseStarted, PluginPhaseStarted)
			fc := tt.args.fakeClient()
			// common setup code
			mockResourceHandler := &pluginsk8sMock.Plugin{}
			mockResourceHandler.OnGetProperties().Return(k8s.PluginProperties{})
			mockResourceHandler.On("BuildIdentityResource", mock.Anything, tctx.TaskExecutionMetadata()).Return(&v1.Pod{}, nil)
			mockResourceHandler.On("GetTaskPhase", mock.Anything, mock.Anything, mock.Anything).Return(tt.args.getTaskPhaseCB())
			pluginManager, err := NewPluginManager(ctx, dummySetupContext(fc), k8s.PluginEntry{
				ID:              "x",
				ResourceToWatch: &v1.Pod{},
				Plugin:          mockResourceHandler,
			}, NewResourceMonitorIndex())
			assert.NotNil(t, res)
			assert.NoError(t, err)

			transition, err := pluginManager.Handle(ctx, tctx)
			if tt.want.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NotNil(t, transition)
			transitionInfo := transition.Info()
			assert.NotNil(t, transitionInfo)
			assert.Equal(t, tt.want.wantPhase, transitionInfo.Phase())
			d := tctx.OutputWriter().(*dummyOutputWriter)
			if tt.want.wantOutputReader {
				assert.NotNil(t, d.r)
			} else {
				assert.Nil(t, d.r)
			}
		})
	}
}

func TestPluginManager_Handle_PluginState(t *testing.T) {
	ctx := context.TODO()
	tm := getMockTaskExecutionMetadata()
	res := &v1.Pod{
		ObjectMeta: v12.ObjectMeta{
			Name:      tm.GetTaskExecutionID().GetGeneratedName(),
			Namespace: tm.GetNamespace(),
		},
	}

	pluginStateQueued := PluginState{
		Phase: PluginPhaseStarted,
		K8sPluginState: k8s.PluginState{
			Phase:        pluginsCore.PhaseQueued,
			PhaseVersion: 0,
			Reason:       "foo",
		},
	}
	pluginStateQueuedVersion1 := PluginState{
		Phase: PluginPhaseStarted,
		K8sPluginState: k8s.PluginState{
			Phase:        pluginsCore.PhaseQueued,
			PhaseVersion: 1,
			Reason:       "foo",
		},
	}
	pluginStateQueuedReasonBar := PluginState{
		Phase: PluginPhaseStarted,
		K8sPluginState: k8s.PluginState{
			Phase:        pluginsCore.PhaseQueued,
			PhaseVersion: 0,
			Reason:       "bar",
		},
	}
	pluginStateRunning := PluginState{
		Phase: PluginPhaseStarted,
		K8sPluginState: k8s.PluginState{
			Phase:        pluginsCore.PhaseRunning,
			PhaseVersion: 0,
			Reason:       "",
		},
	}

	phaseInfoQueued := pluginsCore.PhaseInfoQueuedWithTaskInfo(pluginStateQueued.K8sPluginState.PhaseVersion, pluginStateQueued.K8sPluginState.Reason, nil)
	phaseInfoQueuedVersion1 := pluginsCore.PhaseInfoQueuedWithTaskInfo(
		pluginStateQueuedVersion1.K8sPluginState.PhaseVersion,
		pluginStateQueuedVersion1.K8sPluginState.Reason,
		nil,
	)
	phaseInfoQueuedReasonBar := pluginsCore.PhaseInfoQueuedWithTaskInfo(
		pluginStateQueuedReasonBar.K8sPluginState.PhaseVersion,
		pluginStateQueuedReasonBar.K8sPluginState.Reason,
		nil,
	)
	phaseInfoRunning := pluginsCore.PhaseInfoRunning(0, nil)

	tests := []struct {
		name                string
		startPluginState    PluginState
		reportedPhaseInfo   pluginsCore.PhaseInfo
		expectedPluginState PluginState
	}{
		{
			"NoChange",
			pluginStateQueued,
			phaseInfoQueued,
			pluginStateQueued,
		},
		{
			"K8sPhaseChange",
			pluginStateQueued,
			phaseInfoRunning,
			pluginStateRunning,
		},
		{
			"PhaseVersionChange",
			pluginStateQueued,
			phaseInfoQueuedVersion1,
			pluginStateQueuedVersion1,
		},
		{
			"ReasonChange",
			pluginStateQueued,
			phaseInfoQueuedReasonBar,
			pluginStateQueuedReasonBar,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock TaskExecutionContext
			tCtx := &pluginsCoreMock.TaskExecutionContext{}
			tCtx.OnTaskExecutionMetadata().Return(getMockTaskExecutionMetadata())

			tReader := &pluginsCoreMock.TaskReader{}
			tReader.OnReadMatch(mock.Anything).Return(&core.TaskTemplate{}, nil)
			tCtx.OnTaskReader().Return(tReader)

			// mock state reader / writer to use local pluginState variable
			pluginState := &tt.startPluginState
			customStateReader := &pluginsCoreMock.PluginStateReader{}
			customStateReader.OnGetMatch(mock.MatchedBy(func(i interface{}) bool {
				ps, ok := i.(*PluginState)
				if ok {
					*ps = *pluginState
					return true
				}
				return false
			})).Return(uint8(0), nil)
			tCtx.OnPluginStateReader().Return(customStateReader)

			customStateWriter := &pluginsCoreMock.PluginStateWriter{}
			customStateWriter.OnPutMatch(mock.Anything, mock.MatchedBy(func(i interface{}) bool {
				ps, ok := i.(*PluginState)
				if ok {
					*pluginState = *ps
				}
				return ok
			})).Return(nil)
			tCtx.OnPluginStateWriter().Return(customStateWriter)
			tCtx.OnOutputWriter().Return(&dummyOutputWriter{})

			fc := extendedFakeClient{Client: fake.NewFakeClient(res)}

			mockResourceHandler := &pluginsk8sMock.Plugin{}
			mockResourceHandler.OnGetProperties().Return(k8s.PluginProperties{})
			mockResourceHandler.On("BuildIdentityResource", mock.Anything, tCtx.TaskExecutionMetadata()).Return(&v1.Pod{}, nil)
			mockResourceHandler.On("GetTaskPhase", mock.Anything, mock.Anything, mock.Anything).Return(tt.reportedPhaseInfo, nil)

			// create new PluginManager
			pluginManager, err := NewPluginManager(ctx, dummySetupContext(fc), k8s.PluginEntry{
				ID:              "x",
				ResourceToWatch: &v1.Pod{},
				Plugin:          mockResourceHandler,
			}, NewResourceMonitorIndex())
			assert.NoError(t, err)

			// handle plugin
			_, err = pluginManager.Handle(ctx, tCtx)
			assert.NoError(t, err)

			// verify expected PluginState
			newPluginState := PluginState{}
			_, err = tCtx.PluginStateReader().Get(&newPluginState)
			assert.NoError(t, err)

			assert.True(t, reflect.DeepEqual(newPluginState, tt.expectedPluginState))
		})
	}
}

func TestPluginManager_CustomKubeClient(t *testing.T) {
	ctx := context.TODO()
	tctx := getMockTaskContext(PluginPhaseNotStarted, PluginPhaseStarted)
	// common setup code
	mockResourceHandler := &pluginsk8sMock.Plugin{}
	mockResourceHandler.OnGetProperties().Return(k8s.PluginProperties{})
	mockResourceHandler.On("BuildResource", mock.Anything, tctx).Return(&v1.Pod{}, nil)
	fakeClient := fake.NewClientBuilder().Build()
	newFakeClient := &pluginsCoreMock.KubeClient{}
	newFakeClient.On("GetCache").Return(&mocks.FakeInformers{})
	pluginManager, err := NewPluginManager(ctx, dummySetupContext(fakeClient), k8s.PluginEntry{
		ID:              "x",
		ResourceToWatch: &v1.Pod{},
		Plugin:          mockResourceHandler,
		CustomKubeClient: func(ctx context.Context) (pluginsCore.KubeClient, error) {
			return newFakeClient, nil
		},
	}, NewResourceMonitorIndex())
	assert.NoError(t, err)

	assert.Equal(t, newFakeClient, pluginManager.kubeClient)
}

func TestPluginManager_AddObjectMetadata(t *testing.T) {
	genName := "gen-name"
	ns := "ns"
	or := v12.OwnerReference{}
	l := map[string]string{"l1": "lv1"}
	a := map[string]string{"aKey": "aVal"}
	tm := getMockTaskExecutionMetadataCustom(genName, ns, a, l, or)

	cfg := config.GetK8sPluginConfig()

	t.Run("default", func(t *testing.T) {
		o := &v1.Pod{}
		p := pluginsk8sMock.Plugin{}
		p.OnGetProperties().Return(k8s.PluginProperties{})
		pluginManager := PluginManager{plugin: &p}
		pluginManager.AddObjectMetadata(tm, o, cfg)
		assert.Equal(t, genName, o.GetName())
		assert.Equal(t, []v12.OwnerReference{or}, o.GetOwnerReferences())
		assert.Equal(t, ns, o.GetNamespace())
		assert.Equal(t, map[string]string{
			"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
			"aKey": "aVal",
		}, o.GetAnnotations())
		assert.Equal(t, l, o.GetLabels())
		assert.Equal(t, 0, len(o.GetFinalizers()))
	})

	t.Run("Disable OwnerReferences injection", func(t *testing.T) {
		p := pluginsk8sMock.Plugin{}
		p.OnGetProperties().Return(k8s.PluginProperties{DisableInjectOwnerReferences: true})
		pluginManager := PluginManager{plugin: &p}
		o := &v1.Pod{}
		pluginManager.AddObjectMetadata(tm, o, cfg)
		assert.Equal(t, genName, o.GetName())
		// empty OwnerReference since we are ignoring
		assert.Equal(t, 0, len(o.GetOwnerReferences()))
		assert.Equal(t, ns, o.GetNamespace())
		assert.Equal(t, map[string]string{
			"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
			"aKey": "aVal",
		}, o.GetAnnotations())
		assert.Equal(t, l, o.GetLabels())
		assert.Equal(t, 0, len(o.GetFinalizers()))
	})

	t.Run("Disable enabled InjectFinalizer", func(t *testing.T) {
		p := pluginsk8sMock.Plugin{}
		p.OnGetProperties().Return(k8s.PluginProperties{DisableInjectFinalizer: true})
		pluginManager := PluginManager{plugin: &p}
		// enable finalizer injection
		cfg.InjectFinalizer = true
		o := &v1.Pod{}
		pluginManager.AddObjectMetadata(tm, o, cfg)
		assert.Equal(t, genName, o.GetName())
		// empty OwnerReference since we are ignoring
		assert.Equal(t, 1, len(o.GetOwnerReferences()))
		assert.Equal(t, ns, o.GetNamespace())
		assert.Equal(t, map[string]string{
			"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
			"aKey": "aVal",
		}, o.GetAnnotations())
		assert.Equal(t, l, o.GetLabels())
		assert.Equal(t, 0, len(o.GetFinalizers()))
	})

	t.Run("Disable disabled InjectFinalizer", func(t *testing.T) {
		p := pluginsk8sMock.Plugin{}
		p.OnGetProperties().Return(k8s.PluginProperties{DisableInjectFinalizer: true})
		pluginManager := PluginManager{plugin: &p}
		// disable finalizer injection
		cfg.InjectFinalizer = false
		o := &v1.Pod{}
		pluginManager.AddObjectMetadata(tm, o, cfg)
		assert.Equal(t, genName, o.GetName())
		// empty OwnerReference since we are ignoring
		assert.Equal(t, 1, len(o.GetOwnerReferences()))
		assert.Equal(t, ns, o.GetNamespace())
		assert.Equal(t, map[string]string{
			"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
			"aKey": "aVal",
		}, o.GetAnnotations())
		assert.Equal(t, l, o.GetLabels())
		assert.Equal(t, 0, len(o.GetFinalizers()))
	})

}

func TestResourceManagerConstruction(t *testing.T) {
	ctx := context.Background()
	fakeKubeClient := mocks.NewFakeKubeClient()

	scope := promutils.NewScope("test:plugin_manager")
	index := NewResourceMonitorIndex()
	gvk, err := getPluginGvk(&v1.Pod{})
	assert.NoError(t, err)
	assert.Equal(t, gvk.Kind, "Pod")
	si, err := getPluginSharedInformer(ctx, fakeKubeClient, &v1.Pod{})
	assert.NotNil(t, si)
	assert.NoError(t, err)
	rm := index.GetOrCreateResourceLevelMonitor(ctx, scope, si, gvk)
	assert.NotNil(t, rm)
}

func TestFinalize(t *testing.T) {
	t.Run("DeleteResourceOnFinalize=True", func(t *testing.T) {
		ctx := context.Background()
		sCtx := &pluginsCoreMock.SetupContext{}
		fakeKubeClient := mocks.NewFakeKubeClient()
		sCtx.OnKubeClient().Return(fakeKubeClient)

		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{DeleteResourceOnFinalize: true}))
		p := pluginsk8sMock.Plugin{}
		p.OnGetProperties().Return(k8s.PluginProperties{DisableInjectFinalizer: true})
		tctx := getMockTaskContext(PluginPhaseStarted, PluginPhaseStarted)
		o := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tctx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(),
				Namespace: tctx.TaskExecutionMetadata().GetNamespace(),
			},
		}

		assert.NoError(t, fakeKubeClient.GetClient().Create(ctx, o))

		p.OnBuildIdentityResource(ctx, tctx.TaskExecutionMetadata()).Return(o, nil)
		pluginManager := PluginManager{plugin: &p, kubeClient: fakeKubeClient}
		actualO := &v1.Pod{}
		// Assert the object exists before calling finalize
		assert.NoError(t, fakeKubeClient.GetClient().Get(ctx, k8stypes.NamespacedName{
			Name:      o.Name,
			Namespace: o.Namespace,
		}, actualO))

		// Finalize should now delete the object
		assert.NoError(t, pluginManager.Finalize(ctx, tctx))

		// Assert the object is now deleted.
		assert.Error(t, fakeKubeClient.GetClient().Get(ctx, k8stypes.NamespacedName{
			Name:      o.Name,
			Namespace: o.Namespace,
		}, actualO))
	})

	t.Run("DeleteResourceOnFinalize=False", func(t *testing.T) {
		ctx := context.Background()
		sCtx := &pluginsCoreMock.SetupContext{}
		fakeKubeClient := mocks.NewFakeKubeClient()
		sCtx.OnKubeClient().Return(fakeKubeClient)

		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{DeleteResourceOnFinalize: false}))
		p := pluginsk8sMock.Plugin{}
		p.OnGetProperties().Return(k8s.PluginProperties{DisableInjectFinalizer: true})
		tctx := getMockTaskContext(PluginPhaseStarted, PluginPhaseStarted)
		o := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tctx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(),
				Namespace: tctx.TaskExecutionMetadata().GetNamespace(),
			},
		}

		assert.NoError(t, fakeKubeClient.GetClient().Create(ctx, o))

		p.OnBuildIdentityResource(ctx, tctx.TaskExecutionMetadata()).Return(o, nil)
		pluginManager := PluginManager{plugin: &p, kubeClient: fakeKubeClient}
		actualO := &v1.Pod{}
		// Assert the object exists before calling finalize
		assert.NoError(t, fakeKubeClient.GetClient().Get(ctx, k8stypes.NamespacedName{
			Name:      o.Name,
			Namespace: o.Namespace,
		}, actualO))

		// Finalize should now delete the object
		assert.NoError(t, pluginManager.Finalize(ctx, tctx))

		// Assert the object is still here.
		assert.NoError(t, fakeKubeClient.GetClient().Get(ctx, k8stypes.NamespacedName{
			Name:      o.Name,
			Namespace: o.Namespace,
		}, actualO))
	})

}

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey)
}
