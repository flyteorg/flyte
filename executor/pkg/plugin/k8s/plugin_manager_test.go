package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	pluginsCoreMock "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	k8sMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"
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
