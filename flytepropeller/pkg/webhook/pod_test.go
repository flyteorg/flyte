package webhook

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd/api/latest"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/webhook/mocks"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

var (
	expectedSecretsLabelSelector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			secrets.PodLabel: secrets.PodLabelValue,
		},
	}
	expectedImageBuilderLabelSelector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			"test-arbitrary-label": "test-arbitrary-value",
		},
	}
	testImageBuilderConfig = config.ImageBuilderConfig{
		Enabled: true,
		HostnameReplacement: config.HostnameReplacement{
			Existing:    "test.existing.hostname",
			Replacement: "test.hostname",
		},
		LabelSelector: expectedImageBuilderLabelSelector,
	}
	secretManagerConfig = config.EmbeddedSecretManagerConfig{
		FileMountInitContainer: config.FileMountInitContainerConfig{
			ContainerName: config.EmbeddedSecretsFileMountInitContainerName,
		},
	}
)

func TestNewPodCreationWebhookConfig_NewPodCreationWebhookConfig(t *testing.T) {

	t.Run("Defaults with Secrets", func(t *testing.T) {
		ctx := context.Background()

		pm, err := NewPodCreationWebhookConfig(ctx, &config.Config{
			CertDir:                     "testdata",
			ServiceName:                 "my-service",
			EmbeddedSecretManagerConfig: secretManagerConfig,
		}, latest.Scheme, promutils.NewTestScope())

		assert.NoError(t, err)
		assert.NotNil(t, pm)
		assert.Equal(t, 1, len(pm.httpHandlers))
		secretsHTTPHandler := pm.httpHandlers[0]
		assert.Equal(t, expectedSecretsLabelSelector, *secretsHTTPHandler.mutator.LabelSelector())
	})

	t.Run("With additional Image Builder config enabled", func(t *testing.T) {
		ctx := context.Background()

		pm, err := NewPodCreationWebhookConfig(ctx, &config.Config{
			CertDir:            "testdata",
			ServiceName:        "my-service",
			ImageBuilderConfig: testImageBuilderConfig,
		}, latest.Scheme, promutils.NewTestScope())

		assert.NoError(t, err)
		assert.NotNil(t, pm)
		assert.Equal(t, 2, len(pm.httpHandlers))
		secretsHTTPHandler := pm.httpHandlers[0]
		assert.Equal(t, expectedSecretsLabelSelector, *secretsHTTPHandler.mutator.LabelSelector())
		imageBuilderHTTPHandler := pm.httpHandlers[1]
		assert.Equal(t, expectedImageBuilderLabelSelector, *imageBuilderHTTPHandler.mutator.LabelSelector())
	})
}

func TestPodMutator_Mutate(t *testing.T) {
	inputPod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "container1",
				},
			},
		},
	}

	successMutator := &mocks.PodMutator{}
	successMutator.OnID().Return("SucceedingMutator")
	successMutator.OnMutateMatch(mock.Anything, mock.Anything).Return(nil, false, nil)

	failedMutator := &mocks.PodMutator{}
	failedMutator.OnID().Return("FailingMutator")
	admissionError := admission.Errored(http.StatusBadRequest, fmt.Errorf("failing mock"))
	failedMutator.OnMutateMatch(mock.Anything, mock.Anything).Return(nil, false, &admissionError)

	t.Run("Required Mutator Succeeded", func(t *testing.T) {
		ctx := context.Background()
		_, changed, err := successMutator.Mutate(ctx, inputPod.DeepCopy())
		assert.Nil(t, err)
		assert.False(t, changed)
	})

	t.Run("Required Mutator Failed", func(t *testing.T) {
		ctx := context.Background()
		_, _, err := failedMutator.Mutate(ctx, inputPod.DeepCopy())
		assert.NotNil(t, err)
	})
}

func Test_CreateMutationWebhookConfiguration(t *testing.T) {
	ctx := context.Background()
	serviceName := "test-service"
	pm, err := NewPodCreationWebhookConfig(ctx, &config.Config{
		CertDir:     "testdata",
		ServiceName: serviceName,
	}, latest.Scheme, promutils.NewTestScope())
	assert.NoError(t, err)

	t.Run("Empty namespace", func(t *testing.T) {
		namespace := ""
		c, err := pm.CreateMutationWebhookConfiguration(namespace)
		assert.NoError(t, err)
		assert.NotNil(t, c)

		assert.Equal(t, 1, len(c.Webhooks))
		assert.Equal(t, "flyte-pod-webhook.flyte.org", c.Webhooks[0].Name)
		assert.Equal(t, serviceName, c.Webhooks[0].ClientConfig.Service.Name)
		assert.Equal(t, namespace, c.Webhooks[0].ClientConfig.Service.Namespace)
		assert.Equal(t, getPodMutatePath(secret.SecretsID), *c.Webhooks[0].ClientConfig.Service.Path)
		assert.Equal(t, 1, len(c.Webhooks[0].ObjectSelector.MatchLabels))
		assert.Equal(t, 0, len(c.Webhooks[0].ObjectSelector.DeepCopy().MatchExpressions))
		assert.Equal(t, expectedSecretsLabelSelector, *c.Webhooks[0].ObjectSelector)
	})

	t.Run("With namespace", func(t *testing.T) {
		c, err := pm.CreateMutationWebhookConfiguration("my-namespace")
		assert.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("With image builder", func(t *testing.T) {
		pm, err := NewPodCreationWebhookConfig(ctx, &config.Config{
			CertDir:            "testdata",
			ServiceName:        serviceName,
			ImageBuilderConfig: testImageBuilderConfig,
		}, latest.Scheme, promutils.NewTestScope())
		assert.NoError(t, err)
		namespace := "test-namespace"
		c, err := pm.CreateMutationWebhookConfiguration(namespace)
		assert.NoError(t, err)
		assert.NotNil(t, c)

		assert.Equal(t, 2, len(c.Webhooks))
		assert.Equal(t, secretsWebhookName, c.Webhooks[0].Name)
		assert.Equal(t, serviceName, c.Webhooks[0].ClientConfig.Service.Name)
		assert.Equal(t, namespace, c.Webhooks[0].ClientConfig.Service.Namespace)
		assert.Equal(t, getPodMutatePath(secret.SecretsID), *c.Webhooks[0].ClientConfig.Service.Path)
		assert.Equal(t, 1, len(c.Webhooks[0].ObjectSelector.MatchLabels))
		assert.Equal(t, 0, len(c.Webhooks[0].ObjectSelector.DeepCopy().MatchExpressions))
		assert.Equal(t, expectedSecretsLabelSelector, *c.Webhooks[0].ObjectSelector)

		assert.Equal(t, getMutatingWebhookName(ImageBuilderV1ID), c.Webhooks[1].Name)
		assert.Equal(t, serviceName, c.Webhooks[1].ClientConfig.Service.Name)
		assert.Equal(t, namespace, c.Webhooks[1].ClientConfig.Service.Namespace)
		assert.Equal(t, getPodMutatePath(ImageBuilderV1ID), *c.Webhooks[1].ClientConfig.Service.Path)
		assert.Equal(t, expectedImageBuilderLabelSelector, *c.Webhooks[1].ObjectSelector)
	})
}

func Test_GetMutatePath(t *testing.T) {
	assert.Equal(t, "/mutate--v1-pod/secrets", getPodMutatePath(secret.SecretsID))
	assert.Equal(t, "/mutate--v1-pod/image-builder", getPodMutatePath(ImageBuilderV1ID))
}

func Test_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("Defaults", func(t *testing.T) {
		pm, err := NewPodCreationWebhookConfig(context.Background(), &config.Config{
			CertDir:     "testdata",
			ServiceName: "my-service",
		}, latest.Scheme, promutils.NewTestScope())
		assert.NoError(t, err)

		mockRegister := &mocks.HTTPHookRegistererIface{}
		wh := &admission.Webhook{Handler: pm.httpHandlers[0]}
		mockRegister.On("Register", "/mutate--v1-pod/secrets", wh)
		err = pm.Register(ctx, mockRegister)
		assert.Nil(t, err)
	})

	t.Run("With Image Builder", func(t *testing.T) {
		pm, err := NewPodCreationWebhookConfig(context.Background(), &config.Config{
			CertDir:            "testdata",
			ServiceName:        "my-service",
			ImageBuilderConfig: testImageBuilderConfig,
		}, latest.Scheme, promutils.NewTestScope())
		assert.NoError(t, err)

		mockRegister := &mocks.HTTPHookRegistererIface{}
		secretWH := &admission.Webhook{Handler: pm.httpHandlers[0]}
		mockRegister.On("Register", getPodMutatePath(secret.SecretsID), secretWH)
		imageBuilderWH := &admission.Webhook{Handler: pm.httpHandlers[1]}
		mockRegister.On("Register", getPodMutatePath(ImageBuilderV1ID), imageBuilderWH)

		err = pm.Register(ctx, mockRegister)
		assert.Nil(t, err)
	})
}

func Test_MutatorConfigHandle(t *testing.T) {
	ctx := context.Background()
	pm, err := NewPodCreationWebhookConfig(ctx, &config.Config{
		CertDir:            "testdata",
		ServiceName:        "my-service",
		ImageBuilderConfig: testImageBuilderConfig,
	}, latest.Scheme, promutils.NewTestScope())
	assert.NoError(t, err)

	req := admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Object: runtime.RawExtension{
				Raw: []byte(`{
	"apiVersion": "v1",
	"kind": "Pod",
	"metadata": {
			"name": "foo",
			"namespace": "default"
	},
	"spec": {
			"containers": [
					{
							"image": "test.hostname/orgs/default/bar:v2",
							"name": "bar"
					}
			]
	}
}`),
			},
			OldObject: runtime.RawExtension{
				Raw: []byte(`{
	"apiVersion": "v1",
	"kind": "Pod",
	"metadata": {
			"name": "foo",
			"namespace": "default"
	},
	"spec": {
			"containers": [
					{
							"image": "test.hostname/orgs/default/bar:v1",
							"name": "bar"
					}
			]
	}
}`),
			},
		},
	}

	assert.Equal(t, 2, len(pm.httpHandlers))
	for _, mutator := range pm.httpHandlers {
		resp := mutator.Handle(context.Background(), req)
		assert.True(t, resp.Allowed)
	}
}
