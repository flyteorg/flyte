package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	coreMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	k8sMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"
)

func TestFinalizeAndAbortSucceedWhenKindIsNotInstalled(t *testing.T) {
	gvk := schema.GroupVersionKind{Group: "kubernetes.dask.org", Version: "v1", Kind: "DaskJob"}
	noMatch := &meta.NoKindMatchError{GroupKind: gvk.GroupKind(), SearchedVersions: []string{gvk.Version}}
	kubeClient := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
		Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
			return noMatch
		},
		Delete: func(context.Context, client.WithWatch, client.Object, ...client.DeleteOption) error {
			return noMatch
		},
	}).Build()

	plugin := k8sMocks.NewPlugin(t)
	resource := &unstructured.Unstructured{}
	resource.SetGroupVersionKind(gvk)
	plugin.On("BuildIdentityResource", mock.Anything, mock.Anything).Return(resource, nil)
	plugin.On("GetProperties").Return(k8s.PluginProperties{}).Maybe()

	kc := coreMocks.NewKubeClient(t)
	kc.On("GetClient").Return(kubeClient)

	id := coreMocks.NewTaskExecutionID(t)
	id.On("GetGeneratedName").Return("r1-a0").Maybe()
	md := coreMocks.NewTaskExecutionMetadata(t)
	md.On("GetNamespace").Return("flyte").Maybe()
	md.On("GetAnnotations").Return(map[string]string{}).Maybe()
	md.On("GetLabels").Return(map[string]string{}).Maybe()
	md.On("GetTaskExecutionID").Return(id).Maybe()
	md.On("GetOwnerReference").Return(metav1.OwnerReference{}).Maybe()
	tCtx := coreMocks.NewTaskExecutionContext(t)
	tCtx.On("TaskExecutionMetadata").Return(md)

	pm := NewPluginManager("dask", plugin, kc)
	assert.NoError(t, pm.Finalize(context.Background(), tCtx))
	assert.NoError(t, pm.Abort(context.Background(), tCtx))
}
