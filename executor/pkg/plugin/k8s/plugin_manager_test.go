package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	coreMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	k8sMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"
)

// metadataMock builds a TaskExecutionMetadata sufficient for addObjectMetadata, returning the
// given generated name. GetTaskExecutionID is wired with .Maybe() since it's only consulted when
// the object has no name yet.
func metadataMock(generatedName string) *coreMocks.TaskExecutionMetadata {
	tID := &coreMocks.TaskExecutionID{}
	tID.EXPECT().GetGeneratedName().Return(generatedName).Maybe()

	meta := &coreMocks.TaskExecutionMetadata{}
	meta.EXPECT().GetTaskExecutionID().Return(tID).Maybe()
	meta.EXPECT().GetNamespace().Return("ns")
	meta.EXPECT().GetAnnotations().Return(nil)
	meta.EXPECT().GetLabels().Return(nil)
	meta.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{})
	return meta
}

// TestAddObjectMetadata_PreservesPluginSetName guards the fix that lets a plugin choose its own
// object name: the clustered (JobSet) plugin bounds its name to keep derived pod names within the
// 63-char limit, and addObjectMetadata must not clobber it with the (longer) generated name.
func TestAddObjectMetadata_PreservesPluginSetName(t *testing.T) {
	plugin := &k8sMocks.Plugin{}
	plugin.EXPECT().GetProperties().Return(k8s.PluginProperties{})
	pm := &PluginManager{plugin: plugin}

	t.Run("plugin-set name is kept", func(t *testing.T) {
		obj := &corev1.Pod{}
		obj.SetName("plugin-chosen-name")
		pm.addObjectMetadata(metadataMock("generated-name"), obj, &config.K8sPluginConfig{})
		assert.Equal(t, "plugin-chosen-name", obj.GetName())
		assert.Equal(t, "ns", obj.GetNamespace())
	})

	t.Run("empty name defaults to generated name", func(t *testing.T) {
		obj := &corev1.Pod{}
		pm.addObjectMetadata(metadataMock("generated-name"), obj, &config.K8sPluginConfig{})
		assert.Equal(t, "generated-name", obj.GetName())
	})
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

	kubeClient := &coreMocks.KubeClient{}
	kubeClient.EXPECT().GetClient().Return(fakeClient)

	plugin := &k8sMocks.Plugin{}
	plugin.EXPECT().GetProperties().Return(k8s.PluginProperties{})
	plugin.EXPECT().BuildResource(mock.Anything, mock.Anything).Return(&corev1.Pod{}, nil)

	tCtx := &coreMocks.TaskExecutionContext{}
	tCtx.EXPECT().TaskExecutionMetadata().Return(metadataMock("too-long-name"))

	pm := NewPluginManager("test", plugin, kubeClient)

	transition, err := pm.launchResource(context.Background(), tCtx)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhasePermanentFailure, transition.Info().Phase())
	assert.Equal(t, "InvalidResource", transition.Info().Err().GetCode())
}
