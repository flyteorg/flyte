package webhook

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	latest "k8s.io/client-go/kubernetes/scheme"

	"github.com/flyteorg/flyte/flytepropeller/pkg/webhook/config"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

func Test_createMutationConfig(t *testing.T) {
	pm := NewPodMutator(&config.Config{
		CertDir:     "testdata",
		ServiceName: "my-service",
	}, latest.Scheme, promutils.NewTestScope())

	t.Run("Labels set when both env vars present", func(t *testing.T) {
		t.Setenv(PodNameEnvVar, "test-pod")
		t.Setenv(PodNamespaceEnvVar, "")

		kubeClient := fake.NewSimpleClientset()
		err := createMutationConfig(context.Background(), kubeClient, pm, "")
		assert.NoError(t, err)

		cfgList, err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().List(context.Background(), metav1.ListOptions{})
		assert.NoError(t, err)
		assert.Len(t, cfgList.Items, 1)

		labels := cfgList.Items[0].Labels
		assert.Equal(t, "flyte-pod-webhook", labels["app.kubernetes.io/managed-by"])
		assert.Equal(t, "test-pod", labels["flyte.org/webhook-pod-name"])
		assert.Equal(t, "", labels["flyte.org/webhook-pod-namespace"])
	})

	t.Run("No labels when POD_NAME not set", func(t *testing.T) {
		os.Unsetenv(PodNameEnvVar)
		os.Unsetenv(PodNamespaceEnvVar)

		kubeClient := fake.NewSimpleClientset()
		err := createMutationConfig(context.Background(), kubeClient, pm, "")
		assert.NoError(t, err)

		cfgList, err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().List(context.Background(), metav1.ListOptions{})
		assert.NoError(t, err)
		assert.Len(t, cfgList.Items, 1)

		labels := cfgList.Items[0].Labels
		assert.Empty(t, labels["app.kubernetes.io/managed-by"])
		assert.Empty(t, labels["flyte.org/webhook-pod-name"])
	})

	t.Run("No labels when POD_NAMESPACE not set", func(t *testing.T) {
		t.Setenv(PodNameEnvVar, "test-pod")
		os.Unsetenv(PodNamespaceEnvVar)

		kubeClient := fake.NewSimpleClientset()
		err := createMutationConfig(context.Background(), kubeClient, pm, "")
		assert.NoError(t, err)

		cfgList, err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().List(context.Background(), metav1.ListOptions{})
		assert.NoError(t, err)
		assert.Len(t, cfgList.Items, 1)

		labels := cfgList.Items[0].Labels
		assert.Empty(t, labels["app.kubernetes.io/managed-by"])
		assert.Empty(t, labels["flyte.org/webhook-pod-name"])
	})
}
