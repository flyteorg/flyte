package k8s

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	pluginsCoreMock "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	k8sMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func TestAddObjectMetadata_GeneratedNameMaxLength(t *testing.T) {
	longName := "a-very-long-task-execution-generated-name-that-exceeds-the-limit-12345"

	t.Run("nil GeneratedNameMaxLength preserves generated name", func(t *testing.T) {
		mockPlugin := k8sMocks.NewPlugin(t)
		mockPlugin.EXPECT().GetProperties().Return(k8s.PluginProperties{
			GeneratedNameMaxLength: nil,
		})

		pm := NewPluginManager("test-plugin", mockPlugin, nil)

		taskMetadata := &pluginsCoreMock.TaskExecutionMetadata{}
		taskMetadata.EXPECT().GetNamespace().Return("test-ns")
		taskMetadata.EXPECT().GetAnnotations().Return(nil)
		taskMetadata.EXPECT().GetLabels().Return(nil)
		taskMetadata.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{})

		tID := &pluginsCoreMock.TaskExecutionID{}
		tID.EXPECT().GetGeneratedName().Return(longName)
		taskMetadata.EXPECT().GetTaskExecutionID().Return(tID)

		pod := &v1.Pod{}
		pm.addObjectMetadata(taskMetadata, pod, &config.K8sPluginConfig{})

		assert.Equal(t, longName, pod.GetName())
		assert.Equal(t, "test-ns", pod.GetNamespace())
	})

	t.Run("GeneratedNameMaxLength set truncates long generated name", func(t *testing.T) {
		maxLength := 20
		mockPlugin := k8sMocks.NewPlugin(t)
		mockPlugin.EXPECT().GetProperties().Return(k8s.PluginProperties{
			GeneratedNameMaxLength: ptr.To(maxLength),
		})

		pm := NewPluginManager("test-plugin", mockPlugin, nil)

		taskMetadata := &pluginsCoreMock.TaskExecutionMetadata{}
		taskMetadata.EXPECT().GetNamespace().Return("test-ns")
		taskMetadata.EXPECT().GetAnnotations().Return(nil)
		taskMetadata.EXPECT().GetLabels().Return(nil)
		taskMetadata.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{})

		tID := &pluginsCoreMock.TaskExecutionID{}
		tID.EXPECT().GetGeneratedName().Return(longName)
		tID.EXPECT().GetGeneratedNameWith(0, maxLength).Return(encoding.FixedLengthUniqueID(longName, maxLength))
		taskMetadata.EXPECT().GetTaskExecutionID().Return(tID)

		pod := &v1.Pod{}
		pm.addObjectMetadata(taskMetadata, pod, &config.K8sPluginConfig{})

		expectedName, err := encoding.FixedLengthUniqueID(longName, maxLength)
		require.NoError(t, err)

		assert.Equal(t, expectedName, pod.GetName())
		assert.LessOrEqual(t, len(pod.GetName()), maxLength)
		assert.Equal(t, "test-ns", pod.GetNamespace())
	})

	t.Run("GeneratedNameMaxLength set with short name preserves name", func(t *testing.T) {
		maxLength := 50
		shortName := "short-name"
		mockPlugin := k8sMocks.NewPlugin(t)
		mockPlugin.EXPECT().GetProperties().Return(k8s.PluginProperties{
			GeneratedNameMaxLength: ptr.To(maxLength),
		})

		pm := NewPluginManager("test-plugin", mockPlugin, nil)

		taskMetadata := &pluginsCoreMock.TaskExecutionMetadata{}
		taskMetadata.EXPECT().GetNamespace().Return("test-ns")
		taskMetadata.EXPECT().GetAnnotations().Return(nil)
		taskMetadata.EXPECT().GetLabels().Return(nil)
		taskMetadata.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{})

		tID := &pluginsCoreMock.TaskExecutionID{}
		tID.EXPECT().GetGeneratedName().Return(shortName)
		tID.EXPECT().GetGeneratedNameWith(0, maxLength).Return(shortName, nil)
		taskMetadata.EXPECT().GetTaskExecutionID().Return(tID)

		pod := &v1.Pod{}
		pm.addObjectMetadata(taskMetadata, pod, &config.K8sPluginConfig{})

		assert.Equal(t, shortName, pod.GetName())
	})
}

// metadataMock builds a TaskExecutionMetadata sufficient for addObjectMetadata, returning the
// given generated name.
func metadataMock(generatedName string) *pluginsCoreMock.TaskExecutionMetadata {
	tID := &pluginsCoreMock.TaskExecutionID{}
	tID.EXPECT().GetGeneratedName().Return(generatedName).Maybe()

	meta := &pluginsCoreMock.TaskExecutionMetadata{}
	meta.EXPECT().GetTaskExecutionID().Return(tID).Maybe()
	meta.EXPECT().GetNamespace().Return("ns")
	meta.EXPECT().GetAnnotations().Return(nil)
	meta.EXPECT().GetLabels().Return(nil)
	meta.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{})
	return meta
}

// TestLaunchResource_InvalidCreateFastFails verifies that a deterministic admission/validation
// rejection (e.g. a derived name exceeding k8s length limits) fast-fails instead of looping via
// UnknownTransition and leaving the execution stuck RUNNING.
func TestLaunchResource_InvalidCreateFastFails(t *testing.T) {
	invalidErr := k8serrors.NewInvalid(
		schema.GroupKind{Kind: "JobSet"}, "too-long-name",
		field.ErrorList{field.Invalid(field.NewPath("metadata", "name"), "x", "must be no more than 63 characters")},
	)
	fakeClient := fake.NewClientBuilder().
		WithScheme(k8sscheme.Scheme).
		WithInterceptorFuncs(interceptor.Funcs{
			Create: func(context.Context, client.WithWatch, client.Object, ...client.CreateOption) error {
				return invalidErr
			},
		}).
		Build()

	kubeClient := &pluginsCoreMock.KubeClient{}
	kubeClient.EXPECT().GetClient().Return(fakeClient)

	plugin := &k8sMocks.Plugin{}
	plugin.EXPECT().GetProperties().Return(k8s.PluginProperties{})
	plugin.EXPECT().BuildResource(mock.Anything, mock.Anything).Return(&v1.Pod{}, nil)

	tCtx := &pluginsCoreMock.TaskExecutionContext{}
	tCtx.EXPECT().TaskExecutionMetadata().Return(metadataMock("too-long-name"))

	pm := NewPluginManager("test", plugin, kubeClient)

	transition, err := pm.launchResource(context.Background(), tCtx)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhasePermanentFailure, transition.Info().Phase())
	assert.Equal(t, "InvalidResource", transition.Info().Err().GetCode())
}

func TestHandle_CorruptedPluginStateFailsPermanently(t *testing.T) {
	stateReader := &pluginsCoreMock.PluginStateReader{}
	// (0, err) matches PluginStateManager's contract: the version is zeroed on decode failure.
	stateReader.EXPECT().Get(mock.Anything).Return(uint8(0), fmt.Errorf("gob: decode failed"))

	tCtx := &pluginsCoreMock.TaskExecutionContext{}
	tCtx.EXPECT().PluginStateReader().Return(stateReader)

	plugin := &k8sMocks.Plugin{}
	plugin.EXPECT().GetProperties().Return(k8s.PluginProperties{})

	pm := NewPluginManager("test", plugin, &pluginsCoreMock.KubeClient{})

	transition, err := pm.Handle(context.Background(), tCtx)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhasePermanentFailure, transition.Info().Phase())
	assert.Equal(t, string(errors.CorruptedPluginState), transition.Info().Err().GetCode())
	assert.Equal(t, core.ExecutionError_SYSTEM, transition.Info().Err().GetKind())
}

// TestAddObjectMetadata_ManagedLabel verifies that the label the manager's Pod cache selects
// on survives addObjectMetadata. NewTaskExecutionMetadata is the single injection point; this
// asserts nothing here drops it, since a Pod without the label is invisible to the executor
// that created it and checkResourcePhase would report it as deleted externally.
func TestAddObjectMetadata_ManagedLabel(t *testing.T) {
	newManager := func(t *testing.T) *PluginManager {
		mockPlugin := k8sMocks.NewPlugin(t)
		mockPlugin.EXPECT().GetProperties().Return(k8s.PluginProperties{})
		return NewPluginManager("test-plugin", mockPlugin, nil)
	}

	// Mirrors what NewTaskExecutionMetadata injects into every task's labels.
	managedLabels := map[string]string{
		flytek8s.ManagedLabelKey: flytek8s.ManagedLabelValue,
	}

	newMeta := func(labels map[string]string) *pluginsCoreMock.TaskExecutionMetadata {
		tID := &pluginsCoreMock.TaskExecutionID{}
		tID.EXPECT().GetGeneratedName().Return("name").Maybe()
		meta := &pluginsCoreMock.TaskExecutionMetadata{}
		meta.EXPECT().GetTaskExecutionID().Return(tID).Maybe()
		meta.EXPECT().GetNamespace().Return("ns")
		meta.EXPECT().GetAnnotations().Return(nil)
		meta.EXPECT().GetLabels().Return(labels)
		meta.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{})
		return meta
	}

	t.Run("propagates to the object", func(t *testing.T) {
		pod := &v1.Pod{}
		newManager(t).addObjectMetadata(newMeta(managedLabels), pod, &config.K8sPluginConfig{})

		assert.Equal(t, flytek8s.ManagedLabelValue, pod.GetLabels()[flytek8s.ManagedLabelKey])
	})

	t.Run("matches the selector the manager cache is configured with", func(t *testing.T) {
		pod := &v1.Pod{}
		newManager(t).addObjectMetadata(newMeta(managedLabels), pod, &config.K8sPluginConfig{})

		selector := labels.SelectorFromSet(labels.Set{
			flytek8s.ManagedLabelKey: flytek8s.ManagedLabelValue,
		})
		assert.True(t, selector.Matches(labels.Set(pod.GetLabels())))
	})

	t.Run("not dropped by platform defaults or labels already on the object", func(t *testing.T) {
		pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{flytek8s.ManagedLabelKey: "object-value"},
		}}
		cfg := &config.K8sPluginConfig{
			DefaultLabels: map[string]string{flytek8s.ManagedLabelKey: "default-value"},
		}
		newManager(t).addObjectMetadata(newMeta(managedLabels), pod, cfg)

		assert.Equal(t, flytek8s.ManagedLabelValue, pod.GetLabels()[flytek8s.ManagedLabelKey])
	})
}

func TestAddObjectMetadata_StampsTaskLabelsOnPod(t *testing.T) {
	taskExecID := pluginsCoreMock.NewTaskExecutionID(t)
	taskExecID.EXPECT().GetGeneratedName().Return("run-name-action-name-0")

	taskMeta := pluginsCoreMock.NewTaskExecutionMetadata(t)
	taskMeta.EXPECT().GetNamespace().Return("project-development")
	taskMeta.EXPECT().GetAnnotations().Return(map[string]string{"flyte/annotation": "value"})
	taskMeta.EXPECT().GetLabels().Return(map[string]string{
		"project":   "project",
		"domain":    "development",
		"run":       "run-name",
		"action":    "action-name",
		"attempt":   "2",
		"task-name": "my_module.my_task",
	})
	taskMeta.EXPECT().GetTaskExecutionID().Return(taskExecID)
	taskMeta.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{Name: "owner"})

	plugin := k8sMocks.NewPlugin(t)
	plugin.EXPECT().GetProperties().Return(k8s.PluginProperties{})

	pm := NewPluginManager("test", plugin, nil)

	pod := &v1.Pod{}
	pm.addObjectMetadata(taskMeta, pod, &config.K8sPluginConfig{
		DefaultLabels: map[string]string{"cluster": "default"},
	})

	assert.Equal(t, map[string]string{
		"cluster":   "default",
		"project":   "project",
		"domain":    "development",
		"run":       "run-name",
		"action":    "action-name",
		"attempt":   "2",
		"task-name": "my_module.my_task",
	}, pod.GetLabels())
	assert.Equal(t, "project-development", pod.GetNamespace())
	assert.Equal(t, "run-name-action-name-0", pod.GetName())
}
