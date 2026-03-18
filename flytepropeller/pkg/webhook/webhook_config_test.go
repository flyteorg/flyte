package webhook

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd/api/latest"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/flyteorg/flyte/flytepropeller/pkg/secret"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/webhook/mocks"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// Use the shared constants from pod_handler_test.go

func TestWebhookConfig_CreateMutationWebhookConfiguration(t *testing.T) {
	ctx := context.Background()
	serviceName := "test-service"

	wc, err := NewWebhookConfig(ctx, &config.Config{
		CertDir:                     "testdata",
		ServiceName:                 serviceName,
		EmbeddedSecretManagerConfig: secretManagerConfig,
	}, latest.Scheme, testNamespace, promutils.NewTestScope(), []byte{})

	assert.NoError(t, err)

	t.Run("Empty namespace", func(t *testing.T) {
		c, err := wc.CreateMutationWebhookConfiguration("")
		assert.NoError(t, err)
		assert.NotNil(t, c)

		assert.Equal(t, 1, len(c.Webhooks))
		assert.Equal(t, secretsWebhookName, c.Webhooks[0].Name)
		assert.Equal(t, serviceName, c.Webhooks[0].ClientConfig.Service.Name)
		assert.Equal(t, "", c.Webhooks[0].ClientConfig.Service.Namespace)
		assert.Equal(t, "/mutate--v1-pod/secrets", *c.Webhooks[0].ClientConfig.Service.Path)
		assert.Equal(t, 1, len(c.Webhooks[0].ObjectSelector.MatchLabels))
		assert.Equal(t, expectedSecretsLabelSelector, *c.Webhooks[0].ObjectSelector)
	})

	t.Run("With namespace", func(t *testing.T) {
		c, err := wc.CreateMutationWebhookConfiguration(testNamespace)
		assert.NoError(t, err)
		assert.NotNil(t, c)

		assert.Equal(t, serviceName, c.Webhooks[0].ClientConfig.Service.Name)
		assert.Equal(t, testNamespace, c.Webhooks[0].ClientConfig.Service.Namespace)
	})

	t.Run("With image builder", func(t *testing.T) {
		wc2, err := NewWebhookConfig(ctx, &config.Config{
			CertDir:                     "testdata",
			ServiceName:                 serviceName,
			ImageBuilderConfig:          testImageBuilderConfig,
			EmbeddedSecretManagerConfig: secretManagerConfig,
		}, latest.Scheme, testNamespace, promutils.NewTestScope(), []byte{})
		assert.NoError(t, err)
		c, err := wc2.CreateMutationWebhookConfiguration(testNamespace)
		assert.NoError(t, err)
		assert.NotNil(t, c)

		assert.Equal(t, 2, len(c.Webhooks))
		assert.Equal(t, secretsWebhookName, c.Webhooks[0].Name)
		assert.Equal(t, serviceName, c.Webhooks[0].ClientConfig.Service.Name)
		assert.Equal(t, testNamespace, c.Webhooks[0].ClientConfig.Service.Namespace)
		assert.Equal(t, "/mutate--v1-pod/secrets", *c.Webhooks[0].ClientConfig.Service.Path)
		assert.Equal(t, 1, len(c.Webhooks[0].ObjectSelector.MatchLabels))
		assert.Equal(t, expectedSecretsLabelSelector, *c.Webhooks[0].ObjectSelector)
		assert.Equal(t, getMutatingWebhookName(ImageBuilderV1ID), c.Webhooks[1].Name)
		assert.Equal(t, serviceName, c.Webhooks[1].ClientConfig.Service.Name)
		assert.Equal(t, testNamespace, c.Webhooks[1].ClientConfig.Service.Namespace)
		assert.Equal(t, expectedImageBuilderLabelSelector, *c.Webhooks[1].ObjectSelector)
	})
}

func TestWebhookConfig_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("Defaults", func(t *testing.T) {
		wc, err := NewWebhookConfig(ctx, &config.Config{
			CertDir:                     "testdata",
			ServiceName:                 "test-service",
			EmbeddedSecretManagerConfig: secretManagerConfig,
		}, latest.Scheme, testNamespace, promutils.NewTestScope(), []byte{})
		assert.NoError(t, err)

		mockRegister := &mocks.HTTPHookRegistererIface{}
		handlers := wc.GetHandlers()
		wh := &admission.Webhook{Handler: handlers[0].GetAdmissionHandler()}
		mockRegister.On("Register", handlers[0].GetPath(), wh)
		err = wc.Register(ctx, mockRegister)
		assert.Nil(t, err)
	})

	t.Run("With Image Builder", func(t *testing.T) {
		wc, err := NewWebhookConfig(ctx, &config.Config{
			CertDir:                     "testdata",
			ServiceName:                 "test-service",
			ImageBuilderConfig:          testImageBuilderConfig,
			EmbeddedSecretManagerConfig: secretManagerConfig,
		}, latest.Scheme, testNamespace, promutils.NewTestScope(), []byte{})
		assert.NoError(t, err)

		mockRegister := &mocks.HTTPHookRegistererIface{}
		handlers := wc.GetHandlers()
		secretWH := &admission.Webhook{Handler: handlers[0].GetAdmissionHandler()}
		mockRegister.On("Register", handlers[0].GetPath(), secretWH)
		imageBuilderWH := &admission.Webhook{Handler: handlers[1].GetAdmissionHandler()}
		mockRegister.On("Register", handlers[1].GetPath(), imageBuilderWH)

		err = wc.Register(ctx, mockRegister)
		assert.Nil(t, err)
	})
}

func TestNewWebhookConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("Basic configuration", func(t *testing.T) {
		cfg := &config.Config{
			ServiceName:                 "test-service",
			EmbeddedSecretManagerConfig: secretManagerConfig,
		}

		caBytes := []byte("test-ca-bytes")
		wc, err := NewWebhookConfig(ctx, cfg, latest.Scheme, testNamespace, promutils.NewTestScope(), caBytes)
		assert.NoError(t, err)
		assert.NotNil(t, wc)

		handlers := wc.GetHandlers()
		assert.GreaterOrEqual(t, len(handlers), 1) // At least secrets handler
	})

	t.Run("With image builder enabled", func(t *testing.T) {
		cfg := &config.Config{
			ServiceName:                 "test-service",
			EmbeddedSecretManagerConfig: secretManagerConfig,
			ImageBuilderConfig:          testImageBuilderConfig,
		}

		caBytes := []byte("test-ca-bytes")
		wc, err := NewWebhookConfig(ctx, cfg, latest.Scheme, testNamespace, promutils.NewTestScope(), caBytes)
		assert.NoError(t, err)
		assert.NotNil(t, wc)

		handlers := wc.GetHandlers()
		assert.GreaterOrEqual(t, len(handlers), 2) // Secrets + image builder handlers
	})
}

func TestWebhookConfig_GetHandlers(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		ServiceName:                 "test-service",
		EmbeddedSecretManagerConfig: secretManagerConfig,
	}

	caBytes := []byte("test-ca-bytes")
	wc, err := NewWebhookConfig(ctx, cfg, latest.Scheme, testNamespace, promutils.NewTestScope(), caBytes)
	assert.NoError(t, err)

	handlers := wc.GetHandlers()
	assert.NotEmpty(t, handlers)

	// Verify that each handler implements ResourceHandler interface
	for _, handler := range handlers {
		assert.NotEmpty(t, handler.GetPath())
		assert.NotNil(t, handler.GetAdmissionHandler())
	}
}

func TestVerifyHandlers(t *testing.T) {
	ctx := context.Background()

	t.Run("No duplicates", func(t *testing.T) {
		cfg := &config.Config{
			ServiceName:                 "test-service",
			EmbeddedSecretManagerConfig: secretManagerConfig,
		}

		podHandlers, _, err := initializePodHandlers(ctx, cfg, latest.Scheme, testNamespace, promutils.NewTestScope())
		assert.NoError(t, err)

		nodeHandlers, err := initializeNodeHandlers(ctx, cfg, latest.Scheme, promutils.NewTestScope())
		assert.NoError(t, err)

		allHandlers := append(podHandlers, nodeHandlers...)
		err = verifyHandlers(ctx, allHandlers)
		assert.NoError(t, err)
	})
}

func TestWebhookConfig_Utilities(t *testing.T) {
	t.Run("GenerateMutatePath", func(t *testing.T) {
		pod := &corev1.Pod{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
		}
		gvk := pod.GroupVersionKind()

		path := generateMutatePath(gvk, "test-subpath")
		expectedPath := "/mutate--v1-pod/test-subpath"
		assert.Equal(t, expectedPath, path)
	})

	t.Run("GetMutatingWebhookName", func(t *testing.T) {
		name := getMutatingWebhookName("test-id")
		expectedName := "test-id-webhook.union.ai"
		assert.Equal(t, expectedName, name)
	})

	t.Run("GetPodMutatePath", func(t *testing.T) {
		assert.Equal(t, "/mutate--v1-pod/secrets", getPodMutatePath(secret.SecretsID))
		assert.Equal(t, "/mutate--v1-pod/image-builder", getPodMutatePath(ImageBuilderV1ID))
	})

	t.Run("GetNodeMutatePath", func(t *testing.T) {
		assert.Equal(t, "/mutate--v1-node/node-startup-taints", getNodeMutatePath(NodeStartupTaintsMutatorID))
		assert.Equal(t, "/mutate--v1-node/test-id", getNodeMutatePath("test-id"))
	})
}

// TestWebhookConfig_PodMutationIntegration tests end-to-end pod mutation through WebhookConfig
func TestWebhookConfig_PodMutationIntegration(t *testing.T) {
	ctx := context.Background()

	wc, err := NewWebhookConfig(ctx, &config.Config{
		CertDir:                     "testdata",
		ServiceName:                 "test-service",
		ImageBuilderConfig:          testImageBuilderConfig,
		EmbeddedSecretManagerConfig: secretManagerConfig,
	}, latest.Scheme, testNamespace, promutils.NewTestScope(), []byte{})

	assert.NoError(t, err)

	req := admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Object: runtime.RawExtension{
				Raw: []byte(`{
	"apiVersion": "v1",
	"kind": "Pod",
	"metadata": {
		"name": "test",
		"namespace": "test-namespace",
		"labels": {
			"flyte.org/inject-flyte-secrets": "true",
			"test-arbitrary-label": "test-arbitrary-value"
		}
	},
	"spec": {
		"containers": [
			{
				"name": "test",
				"image": "test.existing.hostname/unionai/image:latest"
			}
		]
	}
}`),
			},
		},
	}

	// Get pod handlers from the webhook config
	var podHandlers []*PodHandler
	for _, handler := range wc.GetHandlers() {
		if podHandler, ok := handler.(*PodHandler); ok {
			podHandlers = append(podHandlers, podHandler)
		}
	}
	assert.Equal(t, 2, len(podHandlers)) // Secrets + Image Builder

	// Test that all pod handlers can process the request successfully
	for _, mutator := range podHandlers {
		resp := mutator.Handle(context.Background(), req)
		assert.True(t, resp.Allowed)
	}
}

// TestWebhookConfig_NodeStartupTaintsIntegration tests that node startup taints webhook is properly integrated
func TestWebhookConfig_NodeStartupTaintsIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateMutationWebhookConfiguration includes node handlers when available", func(t *testing.T) {
		// Create a mock configuration that would have node handlers
		// Note: Since global config controls node startup taints, we test the general structure
		wc, err := NewWebhookConfig(ctx, &config.Config{
			CertDir:                     "testdata",
			ServiceName:                 "test-service",
			EmbeddedSecretManagerConfig: secretManagerConfig,
			ImageBuilderConfig:          testImageBuilderConfig,
		}, latest.Scheme, testNamespace, promutils.NewTestScope(), []byte{})

		assert.NoError(t, err)

		// Verify that webhook config properly handles both pod and node handlers
		handlers := wc.GetHandlers()

		// Should have at least pod handlers (secrets + image builder)
		assert.GreaterOrEqual(t, len(handlers), 2)

		// Verify that CreateMutationWebhookConfiguration works with all handlers
		c, err := wc.CreateMutationWebhookConfiguration(testNamespace)
		assert.NoError(t, err)
		assert.NotNil(t, c)

		// Verify webhook configuration metadata
		assert.Equal(t, "test-service", c.Name)
		assert.Equal(t, testNamespace, c.Namespace)

		// All handlers should produce webhooks
		assert.Equal(t, len(handlers), len(c.Webhooks))

		// Verify each webhook has proper configuration
		for i, webhook := range c.Webhooks {
			assert.NotEmpty(t, webhook.Name)
			assert.Equal(t, "test-service", webhook.ClientConfig.Service.Name)
			assert.Equal(t, testNamespace, webhook.ClientConfig.Service.Namespace)
			assert.NotNil(t, webhook.ClientConfig.Service.Path)

			// Verify the webhook name follows the expected pattern
			// Note: secrets webhook uses a legacy name, others follow new pattern
			if webhook.Name != secretsWebhookName {
				// New webhook naming pattern (all non-legacy webhooks)
				assert.Contains(t, webhook.Name, "-webhook.union.ai")
			}

			// Verify the path corresponds to the handler type
			path := handlers[i].GetPath()
			assert.Equal(t, path, *webhook.ClientConfig.Service.Path)
		}
	})

	t.Run("Webhook name generation for node startup taints", func(t *testing.T) {
		expectedName := getMutatingWebhookName(NodeStartupTaintsMutatorID)
		assert.Equal(t, "node-startup-taints-webhook.union.ai", expectedName)
	})
}
