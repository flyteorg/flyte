package k8s

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"k8s.io/client-go/kubernetes/scheme"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils/labeled"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/task/backoff"
	"github.com/lyft/flytestdlib/promutils"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/cache/informertest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pluginsCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	pluginsCoreMock "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/k8s"
	pluginsk8sMock "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"

	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/lyft/flytepropeller/pkg/controller/executors/mocks"
)

type extendedFakeClient struct {
	client.Client
	CreateError error
	GetError    error
	DeleteError error
}

func (e extendedFakeClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	if e.CreateError != nil {
		return e.CreateError
	}
	return e.Client.Create(ctx, obj)
}

func (e extendedFakeClient) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	if e.GetError != nil {
		return e.GetError
	}
	return e.Client.Get(ctx, key, obj)
}

func (e extendedFakeClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	if e.DeleteError != nil {
		return e.DeleteError
	}

	return e.Client.Delete(ctx, obj, opts...)
}

type k8sSampleHandler struct {
}

func (k8sSampleHandler) BuildResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionContext) (k8s.Resource, error) {
	panic("implement me")
}

func (k8sSampleHandler) BuildIdentityResource(ctx context.Context, taskCtx pluginsCore.TaskExecutionMetadata) (k8s.Resource, error) {
	panic("implement me")
}

func (k8sSampleHandler) GetTaskPhase(ctx context.Context, pluginContext k8s.PluginContext, resource k8s.Resource) (pluginsCore.PhaseInfo, error) {
	panic("implement me")
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
	})
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
	taskExecutionContext.On("TaskExecutionMetadata").Return(getMockTaskExecutionMetadata())

	customStateReader := &pluginsCoreMock.PluginStateReader{}
	customStateReader.On("Get", mock.MatchedBy(func(i interface{}) bool {
		ps, ok := i.(*PluginState)
		if ok {
			ps.Phase = initPhase
			return true
		}
		return false
	})).Return(uint8(0), nil)
	taskExecutionContext.On("PluginStateReader").Return(customStateReader)

	customStateWriter := &pluginsCoreMock.PluginStateWriter{}
	customStateWriter.On("Put", mock.Anything, mock.MatchedBy(func(i interface{}) bool {
		ps, ok := i.(*PluginState)
		return ok && ps.Phase == wantPhase
	})).Return(nil)
	taskExecutionContext.On("PluginStateWriter").Return(customStateWriter)
	taskExecutionContext.On("OutputWriter").Return(&dummyOutputWriter{})

	taskExecutionContext.On("DataStore").Return(nil)
	taskExecutionContext.On("MaxDatasetSizeBytes").Return(int64(0))
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

func dummySetupContext(fakeClient client.Client) pluginsCore.SetupContext {
	setupContext := &pluginsCoreMock.SetupContext{}
	var enqueueOwnerFunc = pluginsCore.EnqueueOwner(func(ownerId k8stypes.NamespacedName) error { return nil })
	setupContext.On("EnqueueOwner").Return(enqueueOwnerFunc)

	kubeClient := &pluginsCoreMock.KubeClient{}
	kubeClient.On("GetClient").Return(fakeClient)
	kubeClient.On("GetCache").Return(&informertest.FakeInformers{})
	setupContext.On("KubeClient").Return(kubeClient)

	setupContext.On("OwnerKind").Return("x")
	setupContext.On("MetricsScope").Return(promutils.NewTestScope())

	return setupContext
}

func TestK8sTaskExecutor_Handle_LaunchResource(t *testing.T) {
	ctx := context.TODO()
	/*var tmpl *core.TaskTemplate
	var inputs *core.LiteralMap*/

	t.Run("jobQueued", func(t *testing.T) {
		tctx := getMockTaskContext(PluginPhaseNotStarted, PluginPhaseStarted)
		// common setup code
		mockResourceHandler := &pluginsk8sMock.Plugin{}
		mockResourceHandler.On("BuildResource", mock.Anything, tctx).Return(&v1.Pod{}, nil)
		fakeClient := fake.NewFakeClient()
		pluginManager, err := NewPluginManager(ctx, dummySetupContext(fakeClient), k8s.PluginEntry{
			ID:              "x",
			ResourceToWatch: &v1.Pod{},
			Plugin:          mockResourceHandler,
		})
		assert.NoError(t, err)

		transition, err := pluginManager.Handle(ctx, tctx)
		assert.NoError(t, err)
		assert.NotNil(t, transition)
		transitionInfo := transition.Info()
		assert.NotNil(t, transitionInfo)
		assert.Equal(t, pluginsCore.PhaseQueued, transitionInfo.Phase())
		createdPod := &v1.Pod{}

		AddObjectMetadata(tctx.TaskExecutionMetadata(), createdPod, &config.K8sPluginConfig{})
		assert.NoError(t, fakeClient.Get(ctx, k8stypes.NamespacedName{Namespace: tctx.TaskExecutionMetadata().GetNamespace(),
			Name: tctx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()}, createdPod))
		assert.Equal(t, tctx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), createdPod.Name)
		assert.NoError(t, fakeClient.Delete(ctx, createdPod))
	})

	t.Run("jobAlreadyExists", func(t *testing.T) {
		tctx := getMockTaskContext(PluginPhaseNotStarted, PluginPhaseStarted)
		// common setup code
		mockResourceHandler := &pluginsk8sMock.Plugin{}
		mockResourceHandler.On("BuildResource", mock.Anything, tctx).Return(&v1.Pod{}, nil)
		fakeClient := fake.NewFakeClient()
		pluginManager, err := NewPluginManager(ctx, dummySetupContext(fakeClient), k8s.PluginEntry{
			ID:              "x",
			ResourceToWatch: &v1.Pod{},
			Plugin:          mockResourceHandler,
		})
		assert.NoError(t, err)

		createdPod := &v1.Pod{}
		AddObjectMetadata(tctx.TaskExecutionMetadata(), createdPod, &config.K8sPluginConfig{})
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
		mockResourceHandler.On("BuildResource", mock.Anything, tctx).Return(&v1.Pod{}, nil)
		fakeClient := extendedFakeClient{
			Client:      fake.NewFakeClient(),
			CreateError: k8serrors.NewForbidden(schema.GroupResource{}, "", errors.New("exceeded quota")),
		}

		pluginManager, err := NewPluginManager(ctx, dummySetupContext(fakeClient), k8s.PluginEntry{
			ID:              "x",
			ResourceToWatch: &v1.Pod{},
			Plugin:          mockResourceHandler,
		})
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
		mockResourceHandler.On("BuildResource", mock.Anything, tctx).Return(&v1.Pod{}, nil)
		fakeClient := extendedFakeClient{
			Client:      fake.NewFakeClient(),
			CreateError: k8serrors.NewForbidden(schema.GroupResource{}, "", errors.New("auth error")),
		}

		pluginManager, err := NewPluginManager(ctx, dummySetupContext(fakeClient), k8s.PluginEntry{
			ID:              "x",
			ResourceToWatch: &v1.Pod{},
			Plugin:          mockResourceHandler,
		})
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
		mockResourceHandler.On("BuildResource", mock.Anything, tctx).Return(&v1.Pod{
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
			Client: fake.NewFakeClient(),
			CreateError: k8serrors.NewForbidden(schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: limits.memory=3Gi, "+
				"used: limits.memory=7976Gi, limited: limits.memory=8000Gi")),
		}

		backOffController := backoff.NewController(ctx)
		pluginManager, err := NewPluginManagerWithBackOff(ctx, dummySetupContext(fakeClient), k8s.PluginEntry{
			ID:              "x",
			ResourceToWatch: &v1.Pod{},
			Plugin:          mockResourceHandler,
		}, backOffController)

		assert.NoError(t, err)
		transition, err := pluginManager.Handle(ctx, tctx)
		assert.NoError(t, err)
		assert.NotNil(t, transition)
		assert.Equal(t, transition.Info().Phase(), pluginsCore.PhaseWaitingForResources)

		// Build a reference resource that is supposed to be identical to the resource built by pluginManager
		referenceResource, err := mockResourceHandler.BuildResource(ctx, tctx)
		assert.NoError(t, err)
		AddObjectMetadata(tctx.TaskExecutionMetadata(), referenceResource, config.GetK8sPluginConfig())
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
		mockResourceHandler.OnBuildIdentityResourceMatch(mock.Anything, tctx.TaskExecutionMetadata()).Return(&v1.Pod{}, nil)
		mockResourceHandler.OnGetTaskPhaseMatch(mock.Anything, mock.Anything, mock.Anything).Return(pluginsCore.PhaseInfo{}, nil)
		pluginManager, err := NewPluginManager(ctx, dummySetupContext(fc), k8s.PluginEntry{
			ID:              "x",
			ResourceToWatch: &v1.Pod{},
			Plugin:          mockResourceHandler,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		err = pluginManager.Abort(ctx, tctx)
		assert.NoError(t, err)
	})

	t.Run("Abort Pod doesn't exist", func(t *testing.T) {
		// common setup code
		tctx := getMockTaskContext(PluginPhaseStarted, PluginPhaseStarted)
		fc := extendedFakeClient{Client: fake.NewFakeClientWithScheme(scheme.Scheme)}
		// common setup code
		mockResourceHandler := &pluginsk8sMock.Plugin{}
		mockResourceHandler.OnBuildIdentityResourceMatch(mock.Anything, tctx.TaskExecutionMetadata()).Return(&v1.Pod{}, nil)
		mockResourceHandler.OnGetTaskPhaseMatch(mock.Anything, mock.Anything, mock.Anything).Return(pluginsCore.PhaseInfo{}, nil)
		pluginManager, err := NewPluginManager(ctx, dummySetupContext(fc), k8s.PluginEntry{
			ID:              "x",
			ResourceToWatch: &v1.Pod{},
			Plugin:          mockResourceHandler,
		})
		assert.NotNil(t, res)
		assert.NoError(t, err)

		err = pluginManager.Abort(ctx, tctx)
		assert.NoError(t, err)
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
			mockResourceHandler.On("BuildIdentityResource", mock.Anything, tctx.TaskExecutionMetadata()).Return(&v1.Pod{}, nil)
			mockResourceHandler.On("GetTaskPhase", mock.Anything, mock.Anything, mock.Anything).Return(tt.args.getTaskPhaseCB())
			pluginManager, err := NewPluginManager(ctx, dummySetupContext(fc), k8s.PluginEntry{
				ID:              "x",
				ResourceToWatch: &v1.Pod{},
				Plugin:          mockResourceHandler,
			})
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

func TestAddObjectMetadata(t *testing.T) {
	genName := "genName"
	execID := &pluginsCoreMock.TaskExecutionID{}
	execID.On("GetGeneratedName").Return(genName)
	tm := &pluginsCoreMock.TaskExecutionMetadata{}
	tm.On("GetTaskExecutionID").Return(execID)
	or := v12.OwnerReference{}
	tm.On("GetOwnerReference").Return(or)
	ns := "ns"
	tm.On("GetNamespace").Return(ns)
	tm.On("GetAnnotations").Return(map[string]string{"aKey": "aVal"})

	l := map[string]string{
		"l1": "lv1",
	}
	tm.On("GetLabels").Return(l)

	o := &v1.Pod{}
	cfg := config.GetK8sPluginConfig()
	AddObjectMetadata(tm, o, cfg)
	assert.Equal(t, genName, o.GetName())
	assert.Equal(t, []v12.OwnerReference{or}, o.GetOwnerReferences())
	assert.Equal(t, ns, o.GetNamespace())
	assert.Equal(t, map[string]string{
		"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
		"aKey": "aVal",
	}, o.GetAnnotations())
	assert.Equal(t, l, o.GetLabels())
}

func init() {
	labeled.SetMetricKeys(contextutils.NamespaceKey)
}
