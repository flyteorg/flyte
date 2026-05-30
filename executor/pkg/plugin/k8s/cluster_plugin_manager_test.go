package k8s

import (
	"context"
	"testing"
	"time"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	coremocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
)

// fakeClusterPlugin is a hand-written k8s.ClusterPlugin used to drive the ClusterPluginManager state
// machine. The clusterReady and jobPhase toggles control IsClusterReady and GetJobPhase responses.
type fakeClusterPlugin struct {
	clusterReady bool
	jobPhase     pluginsCore.Phase
}

var _ k8s.ClusterPlugin = &fakeClusterPlugin{}

func (f *fakeClusterPlugin) GetClusterName(_ context.Context, _ pluginsCore.TaskExecutionContext) (string, error) {
	return "shared-cluster", nil
}

func rayClusterShell() *rayv1.RayCluster {
	return &rayv1.RayCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RayCluster",
			APIVersion: rayv1.SchemeGroupVersion.String(),
		},
	}
}

func rayJobShell() *rayv1.RayJob {
	return &rayv1.RayJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RayJob",
			APIVersion: rayv1.SchemeGroupVersion.String(),
		},
	}
}

func (f *fakeClusterPlugin) BuildClusterResource(_ context.Context, _ pluginsCore.TaskExecutionContext) (client.Object, error) {
	return rayClusterShell(), nil
}

func (f *fakeClusterPlugin) BuildClusterIdentityResource(_ context.Context, _ pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	return rayClusterShell(), nil
}

func (f *fakeClusterPlugin) IsClusterReady(_ context.Context, _ k8s.PluginContext, _ client.Object) (bool, error) {
	return f.clusterReady, nil
}

func (f *fakeClusterPlugin) BuildJobResource(_ context.Context, _ pluginsCore.TaskExecutionContext, clusterName string) (client.Object, error) {
	job := rayJobShell()
	job.Spec = rayv1.RayJobSpec{
		ClusterSelector: map[string]string{"ray.io/cluster": clusterName},
	}
	return job, nil
}

func (f *fakeClusterPlugin) BuildJobIdentityResource(_ context.Context, _ pluginsCore.TaskExecutionMetadata) (client.Object, error) {
	return rayJobShell(), nil
}

func (f *fakeClusterPlugin) GetJobPhase(_ context.Context, _ k8s.PluginContext, _ client.Object) (pluginsCore.PhaseInfo, error) {
	switch f.jobPhase {
	case pluginsCore.PhaseSuccess:
		return pluginsCore.PhaseInfoSuccess(nil), nil
	case pluginsCore.PhaseRunning:
		return pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, nil), nil
	default:
		return pluginsCore.PhaseInfoQueued(time.Now(), pluginsCore.DefaultPhaseVersion, "queued"), nil
	}
}

func (f *fakeClusterPlugin) GetProperties() k8s.PluginProperties {
	return k8s.PluginProperties{}
}

// newFakeKubeClient builds a pluginsCore.KubeClient backed by a controller-runtime fake client that is
// seeded with objs and aware of the ray CRD types.
func newFakeKubeClient(t *testing.T, objs ...client.Object) pluginsCore.KubeClient {
	require.NoError(t, rayv1.AddToScheme(scheme.Scheme))
	c := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(objs...).Build()
	kc := &coremocks.KubeClient{}
	kc.EXPECT().GetClient().Return(c).Maybe()
	kc.EXPECT().GetCache().Return(nil).Maybe()
	return kc
}

// stateCapture records the last plugin state written by the manager.
type stateCapture struct {
	state ClusterPluginState
}

// taskContextForState returns a TaskExecutionContext whose PluginStateReader returns the provided
// initial state and whose PluginStateWriter records the written state into the returned stateCapture.
func taskContextForState(t *testing.T, state ClusterPluginState) (*coremocks.TaskExecutionContext, *stateCapture) {
	capture := &stateCapture{}

	tCtx := &coremocks.TaskExecutionContext{}

	reader := &coremocks.PluginStateReader{}
	reader.EXPECT().Get(mock.Anything).RunAndReturn(func(v interface{}) (uint8, error) {
		if ptr, ok := v.(*ClusterPluginState); ok {
			*ptr = state
		}
		return 0, nil
	}).Maybe()
	tCtx.EXPECT().PluginStateReader().Return(reader).Maybe()

	writer := &coremocks.PluginStateWriter{}
	writer.EXPECT().Put(mock.Anything, mock.Anything).RunAndReturn(func(_ uint8, v interface{}) error {
		if ptr, ok := v.(*ClusterPluginState); ok {
			capture.state = *ptr
		}
		return nil
	}).Maybe()
	tCtx.EXPECT().PluginStateWriter().Return(writer).Maybe()

	tID := &coremocks.TaskExecutionID{}
	tID.EXPECT().GetGeneratedName().Return("job-name").Maybe()

	md := &coremocks.TaskExecutionMetadata{}
	md.EXPECT().GetNamespace().Return("ns").Maybe()
	md.EXPECT().GetTaskExecutionID().Return(tID).Maybe()
	md.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{}).Maybe()
	md.EXPECT().GetLabels().Return(map[string]string{}).Maybe()
	md.EXPECT().GetAnnotations().Return(map[string]string{}).Maybe()
	tCtx.EXPECT().TaskExecutionMetadata().Return(md).Maybe()

	// DataStore / OutputWriter are only reached on the success path; add Maybe stubs so accidental
	// access does not panic.
	tCtx.EXPECT().OutputWriter().Return(nil).Maybe()
	tCtx.EXPECT().DataStore().Return(nil).Maybe()

	return tCtx, capture
}

func TestClusterPluginManager_NotStarted_CreatesClusterNoOwnerRef(t *testing.T) {
	ctx := context.Background()
	kc := newFakeKubeClient(t)
	plugin := &fakeClusterPlugin{}
	m := NewClusterPluginManager("test-id", plugin, kc)

	tCtx, capture := taskContextForState(t, ClusterPluginState{Phase: ClusterPhaseNotStarted})
	transition, err := m.Handle(ctx, tCtx)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseInitializing, transition.Info().Phase())

	assert.Equal(t, ClusterPhaseClusterWait, capture.state.Phase)
	assert.Equal(t, "shared-cluster", capture.state.ClusterName)

	cluster := &rayv1.RayCluster{}
	require.NoError(t, kc.GetClient().Get(ctx, k8stypes.NamespacedName{Namespace: "ns", Name: "shared-cluster"}, cluster))
	assert.Empty(t, cluster.GetOwnerReferences())
	assert.Empty(t, cluster.GetFinalizers())
}

func TestClusterPluginManager_ClusterWait_NotReady_StaysInitializing(t *testing.T) {
	ctx := context.Background()
	existingCluster := rayClusterShell()
	existingCluster.SetNamespace("ns")
	existingCluster.SetName("shared-cluster")
	kc := newFakeKubeClient(t, existingCluster)

	plugin := &fakeClusterPlugin{clusterReady: false}
	m := NewClusterPluginManager("test-id", plugin, kc)

	tCtx, _ := taskContextForState(t, ClusterPluginState{Phase: ClusterPhaseClusterWait, ClusterName: "shared-cluster"})
	transition, err := m.Handle(ctx, tCtx)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseInitializing, transition.Info().Phase())

	jobList := &rayv1.RayJobList{}
	require.NoError(t, kc.GetClient().List(ctx, jobList))
	assert.Empty(t, jobList.Items)
}

func TestClusterPluginManager_ClusterWait_Ready_CreatesJobWithSelector(t *testing.T) {
	ctx := context.Background()
	existingCluster := rayClusterShell()
	existingCluster.SetNamespace("ns")
	existingCluster.SetName("shared-cluster")
	kc := newFakeKubeClient(t, existingCluster)

	plugin := &fakeClusterPlugin{clusterReady: true}
	m := NewClusterPluginManager("test-id", plugin, kc)

	tCtx, capture := taskContextForState(t, ClusterPluginState{Phase: ClusterPhaseClusterWait, ClusterName: "shared-cluster"})
	transition, err := m.Handle(ctx, tCtx)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, transition.Info().Phase())
	assert.Equal(t, ClusterPhaseJobStarted, capture.state.Phase)

	job := &rayv1.RayJob{}
	require.NoError(t, kc.GetClient().Get(ctx, k8stypes.NamespacedName{Namespace: "ns", Name: "job-name"}, job))
	assert.Equal(t, "shared-cluster", job.Spec.ClusterSelector["ray.io/cluster"])
}

func TestClusterPluginManager_JobStarted_MapsRunning(t *testing.T) {
	ctx := context.Background()
	existingJob := rayJobShell()
	existingJob.SetNamespace("ns")
	existingJob.SetName("job-name")
	kc := newFakeKubeClient(t, existingJob)

	plugin := &fakeClusterPlugin{jobPhase: pluginsCore.PhaseRunning}
	m := NewClusterPluginManager("test-id", plugin, kc)

	tCtx, _ := taskContextForState(t, ClusterPluginState{Phase: ClusterPhaseJobStarted, ClusterName: "shared-cluster"})
	transition, err := m.Handle(ctx, tCtx)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, transition.Info().Phase())
}

func TestClusterPluginManager_Abort_DeletesJobNotCluster(t *testing.T) {
	ctx := context.Background()
	existingCluster := rayClusterShell()
	existingCluster.SetNamespace("ns")
	existingCluster.SetName("shared-cluster")
	existingJob := rayJobShell()
	existingJob.SetNamespace("ns")
	existingJob.SetName("job-name")
	kc := newFakeKubeClient(t, existingCluster, existingJob)

	plugin := &fakeClusterPlugin{}
	m := NewClusterPluginManager("test-id", plugin, kc)

	tCtx, _ := taskContextForState(t, ClusterPluginState{Phase: ClusterPhaseJobStarted, ClusterName: "shared-cluster"})
	require.NoError(t, m.Abort(ctx, tCtx))

	job := &rayv1.RayJob{}
	err := kc.GetClient().Get(ctx, k8stypes.NamespacedName{Namespace: "ns", Name: "job-name"}, job)
	assert.True(t, k8serrors.IsNotFound(err), "expected job to be deleted, got err: %v", err)

	cluster := &rayv1.RayCluster{}
	require.NoError(t, kc.GetClient().Get(ctx, k8stypes.NamespacedName{Namespace: "ns", Name: "shared-cluster"}, cluster))
}
