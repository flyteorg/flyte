package k8s

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	coreMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	k8sMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"
)

func newLaunchTestPluginManager(t *testing.T, createErr error) (*PluginManager, pluginsCore.TaskExecutionContext) {
	tID := coreMocks.NewTaskExecutionID(t)
	tID.EXPECT().GetGeneratedName().Return("test-pod")

	meta := coreMocks.NewTaskExecutionMetadata(t)
	meta.EXPECT().GetTaskExecutionID().Return(tID)
	meta.EXPECT().GetNamespace().Return("test-ns")
	meta.EXPECT().GetAnnotations().Return(nil)
	meta.EXPECT().GetLabels().Return(nil)
	meta.EXPECT().GetOwnerReference().Return(metav1.OwnerReference{Kind: "node", Name: "n1"})

	tCtx := coreMocks.NewTaskExecutionContext(t)
	tCtx.EXPECT().TaskExecutionMetadata().Return(meta)

	plugin := k8sMocks.NewPlugin(t)
	plugin.EXPECT().GetProperties().Return(k8s.PluginProperties{})
	plugin.EXPECT().BuildResource(mock.Anything, tCtx).Return(&corev1.Pod{}, nil)

	builder := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme)
	if createErr != nil {
		builder = builder.WithInterceptorFuncs(interceptor.Funcs{
			Create: func(_ context.Context, _ client.WithWatch, _ client.Object, _ ...client.CreateOption) error {
				return createErr
			},
		})
	}

	kubeClient := coreMocks.NewKubeClient(t)
	kubeClient.EXPECT().GetClient().Return(builder.Build())

	return NewPluginManager("test", plugin, kubeClient), tCtx
}

func TestLaunchResourceCreateErrorHandling(t *testing.T) {
	invalidErr := k8serrors.NewInvalid(
		schema.GroupKind{Kind: "Pod"},
		"test-pod",
		field.ErrorList{field.Required(field.NewPath("spec", "containers").Index(0).Child("image"), "")},
	)

	tests := []struct {
		name      string
		createErr error
		wantPhase pluginsCore.Phase
		wantCode  string
		wantKind  core.ExecutionError_ErrorKind
	}{
		{
			name:      "invalid resource is a permanent user failure",
			createErr: invalidErr,
			wantPhase: pluginsCore.PhasePermanentFailure,
			wantCode:  "InvalidResource",
			wantKind:  core.ExecutionError_USER,
		},
		{
			name:      "bad request is a permanent user failure",
			createErr: k8serrors.NewBadRequest("malformed resource"),
			wantPhase: pluginsCore.PhasePermanentFailure,
			wantCode:  "InvalidResource",
			wantKind:  core.ExecutionError_USER,
		},
		{
			name:      "forbidden is a retryable failure",
			createErr: k8serrors.NewForbidden(schema.GroupResource{Resource: "pods"}, "test-pod", errors.New("exceeded quota")),
			wantPhase: pluginsCore.PhaseRetryableFailure,
			wantCode:  "RuntimeFailure",
			wantKind:  core.ExecutionError_USER,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm, tCtx := newLaunchTestPluginManager(t, tt.createErr)

			transition, err := pm.launchResource(t.Context(), tCtx)
			require.NoError(t, err)

			info := transition.Info()
			assert.Equal(t, tt.wantPhase, info.Phase())
			require.NotNil(t, info.Err())
			assert.Equal(t, tt.wantCode, info.Err().GetCode())
			assert.Equal(t, tt.wantKind, info.Err().GetKind())
		})
	}
}

func TestLaunchResourceOtherCreateErrorsAreSystemErrors(t *testing.T) {
	pm, tCtx := newLaunchTestPluginManager(t, k8serrors.NewInternalError(errors.New("etcd unavailable")))

	_, err := pm.launchResource(t.Context(), tCtx)
	require.Error(t, err)
}

func TestLaunchResourceSuccess(t *testing.T) {
	pm, tCtx := newLaunchTestPluginManager(t, nil)

	transition, err := pm.launchResource(t.Context(), tCtx)
	require.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, transition.Info().Phase())
}
