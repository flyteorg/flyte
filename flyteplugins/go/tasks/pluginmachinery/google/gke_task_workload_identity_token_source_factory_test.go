package google

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetGcpServiceAccount(t *testing.T) {
	ctx := context.TODO()

	t.Run("get GCP service account", func(t *testing.T) {
		kubeClient := fake.NewSimpleClientset(&corev1.ServiceAccount{
			ObjectMeta: v1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
				Annotations: map[string]string{
					"owner":                          "abc",
					"iam.gke.io/gcp-service-account": "gcp-service-account",
				},
			}})
		ts := gkeTaskWorkloadIdentityTokenSourceFactory{kubeClient: kubeClient}
		gcpServiceAccount, err := ts.getGcpServiceAccount(ctx, Identity{
			K8sNamespace:      "namespace",
			K8sServiceAccount: "name",
		})

		assert.NoError(t, err)
		assert.Equal(t, "gcp-service-account", gcpServiceAccount)
	})

	t.Run("no GCP service account", func(t *testing.T) {
		kubeClient := fake.NewSimpleClientset()
		ts := gkeTaskWorkloadIdentityTokenSourceFactory{kubeClient: kubeClient}
		_, err := ts.getGcpServiceAccount(ctx, Identity{
			K8sNamespace:      "namespace",
			K8sServiceAccount: "name",
		})

		assert.ErrorContains(t, err, "failed to retrieve task k8s service account")
	})

	t.Run("no GCP service account annotation", func(t *testing.T) {
		kubeClient := fake.NewSimpleClientset(&corev1.ServiceAccount{
			ObjectMeta: v1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
				Annotations: map[string]string{
					"owner": "abc",
				},
			}})
		ts := gkeTaskWorkloadIdentityTokenSourceFactory{kubeClient: kubeClient}
		_, err := ts.getGcpServiceAccount(ctx, Identity{
			K8sNamespace:      "namespace",
			K8sServiceAccount: "name",
		})

		assert.ErrorContains(t, err, "annotation doesn't exist on k8s service account")
	})
}
