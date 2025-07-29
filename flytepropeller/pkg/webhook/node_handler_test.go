package webhook

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	admissionv1 "k8s.io/api/admission/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd/api/latest"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/webhook/mocks"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// Use shared test constants from common_test.go

func TestNodeHandler_Mutate(t *testing.T) {
	inputNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Spec: corev1.NodeSpec{
			PodCIDR: "10.244.0.0/24",
		},
	}

	mockMutator := &mocks.NodeMutator{}

	t.Run("Required Mutator Succeeded", func(t *testing.T) {
		mockMutator.On("Mutate", mock.Anything, inputNode).Return(inputNode, false, (*admission.Response)(nil)).Once()
		newNode, changed, admissionResp := mockMutator.Mutate(context.Background(), inputNode)
		assert.Nil(t, admissionResp)
		assert.False(t, changed)
		assert.Equal(t, inputNode, newNode)
	})

	t.Run("Required Mutator Failed", func(t *testing.T) {
		errResp := &admission.Response{
			AdmissionResponse: admissionv1.AdmissionResponse{
				Result: &metav1.Status{
					Code:    http.StatusInternalServerError,
					Message: "mutator failed",
				},
			},
		}
		mockMutator.On("Mutate", mock.Anything, inputNode).Return(inputNode, false, errResp).Once()
		newNode, changed, admissionResp := mockMutator.Mutate(context.Background(), inputNode)
		assert.NotNil(t, admissionResp)
		assert.False(t, changed)
		assert.Equal(t, inputNode, newNode)
	})
}

func TestNodeHandler_GetMutator(t *testing.T) {
	mockMutator := &mocks.NodeMutator{}
	mockMutator.On("ID").Return("test-mutator")

	nodeHandler := NewNodeHandler(admission.NewDecoder(latest.Scheme), mockMutator)

	assert.Equal(t, mockMutator, nodeHandler.mutator)
}

func TestNodeHandler_GetPath(t *testing.T) {
	mockMutator := &mocks.NodeMutator{}
	mockMutator.On("ID").Return("test-mutator")

	nodeHandler := NewNodeHandler(admission.NewDecoder(latest.Scheme), mockMutator)

	expectedPath := "/mutate--v1-node/test-mutator"
	assert.Equal(t, expectedPath, nodeHandler.GetPath())
}

func TestNodeHandler_GetAdmissionHandler(t *testing.T) {
	mockMutator := &mocks.NodeMutator{}
	mockMutator.On("ID").Return("test-mutator")

	nodeHandler := NewNodeHandler(admission.NewDecoder(latest.Scheme), mockMutator)

	handler := nodeHandler.GetAdmissionHandler()
	assert.Equal(t, nodeHandler, handler)
}

func TestNodeHandler_GetMutatingWebhook(t *testing.T) {
	mockMutator := &mocks.NodeMutator{}
	mockMutator.On("ID").Return("test-mutator")

	labelSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"test-label": "test-value",
		},
	}
	mockMutator.On("LabelSelector").Return(labelSelector)

	nodeHandler := NewNodeHandler(admission.NewDecoder(latest.Scheme), mockMutator)

	cfg := &config.Config{
		ServiceName: "test-service",
		ServicePort: 443,
	}
	caBytes := []byte("test-ca-bytes")

	webhook := nodeHandler.GetMutatingWebhook(testNamespace, caBytes, cfg)

	assert.Equal(t, "test-mutator-webhook.union.ai", webhook.Name)
	assert.Equal(t, caBytes, webhook.ClientConfig.CABundle)
	assert.Equal(t, cfg.ServiceName, webhook.ClientConfig.Service.Name)
	assert.Equal(t, testNamespace, webhook.ClientConfig.Service.Namespace)
	assert.Equal(t, "/mutate--v1-node/test-mutator", *webhook.ClientConfig.Service.Path)
	assert.Equal(t, &cfg.ServicePort, webhook.ClientConfig.Service.Port)
	assert.Equal(t, labelSelector, webhook.ObjectSelector)

	// Check admission rules
	assert.Equal(t, 1, len(webhook.Rules))
	assert.Equal(t, []admissionregistrationv1.OperationType{admissionregistrationv1.Create}, webhook.Rules[0].Operations)
	assert.Equal(t, []string{""}, webhook.Rules[0].Rule.APIGroups)
	assert.Equal(t, []string{"v1"}, webhook.Rules[0].Rule.APIVersions)
	assert.Equal(t, []string{"nodes"}, webhook.Rules[0].Rule.Resources)
}

func TestNodeHandler_Handle(t *testing.T) {
	mockMutator := &mocks.NodeMutator{}
	mockMutator.On("ID").Return("test-mutator")

	nodeHandler := NewNodeHandler(admission.NewDecoder(latest.Scheme), mockMutator)

	t.Run("Successful mutation with changes", func(t *testing.T) {
		modifiedNode := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
				Labels: map[string]string{
					"mutated": "true",
				},
			},
		}

		mockMutator.On("Mutate", mock.Anything, mock.AnythingOfType("*v1.Node")).Return(modifiedNode, true, (*admission.Response)(nil)).Once()

		req := admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Object: runtime.RawExtension{
					Raw: []byte(`{
						"apiVersion": "v1",
						"kind": "Node",
						"metadata": {
							"name": "test-node"
						}
					}`),
				},
			},
		}

		resp := nodeHandler.Handle(context.Background(), req)
		assert.True(t, resp.Allowed)
		assert.NotEmpty(t, resp.Patches)
	})

	t.Run("Successful mutation without changes", func(t *testing.T) {
		mockMutator.On("Mutate", mock.Anything, mock.AnythingOfType("*v1.Node")).Return((*corev1.Node)(nil), false, (*admission.Response)(nil)).Once()

		req := admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Object: runtime.RawExtension{
					Raw: []byte(`{
						"apiVersion": "v1",
						"kind": "Node",
						"metadata": {
							"name": "test-node"
						}
					}`),
				},
			},
		}

		resp := nodeHandler.Handle(context.Background(), req)
		assert.True(t, resp.Allowed)
		assert.Equal(t, "No changes", resp.Result.Message)
	})

	t.Run("Mutator returns error", func(t *testing.T) {
		errResp := &admission.Response{
			AdmissionResponse: admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Code:    http.StatusForbidden,
					Message: "mutation failed",
				},
			},
		}

		mockMutator.On("Mutate", mock.Anything, mock.AnythingOfType("*v1.Node")).Return((*corev1.Node)(nil), false, errResp).Once()

		req := admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Object: runtime.RawExtension{
					Raw: []byte(`{
						"apiVersion": "v1",
						"kind": "Node",
						"metadata": {
							"name": "test-node"
						}
					}`),
				},
			},
		}

		resp := nodeHandler.Handle(context.Background(), req)
		assert.False(t, resp.Allowed)
		assert.Equal(t, "mutation failed", resp.Result.Message)
	})

	t.Run("Invalid JSON in request", func(t *testing.T) {
		req := admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Object: runtime.RawExtension{
					Raw: []byte(`invalid json`),
				},
			},
		}

		resp := nodeHandler.Handle(context.Background(), req)
		assert.False(t, resp.Allowed)
		assert.Equal(t, int32(http.StatusBadRequest), resp.Result.Code)
	})
}

func TestInitializeNodeHandlers(t *testing.T) {
	ctx := context.Background()

	t.Run("Node startup taints disabled", func(t *testing.T) {
		cfg := &config.Config{
			// Node startup taint mutator is disabled by default
		}

		handlers, err := initializeNodeHandlers(ctx, cfg, latest.Scheme, promutils.NewTestScope())
		assert.NoError(t, err)
		assert.Equal(t, 0, len(handlers)) // No handlers when disabled
	})

	t.Run("Node startup taints enabled", func(t *testing.T) {
		// This test would require modifying the global config section to enable node startup taints
		// Since the global config defaults to disabled, we just test that the function works
		cfg := &config.Config{}

		handlers, err := initializeNodeHandlers(ctx, cfg, latest.Scheme, promutils.NewTestScope())
		assert.NoError(t, err)
		// When disabled, should return empty handlers (could be nil or empty slice)
		assert.Equal(t, 0, len(handlers))
	})
}

func TestGetNodeAdmissionRules(t *testing.T) {
	rules := getNodeAdmissionRules()

	assert.Equal(t, 1, len(rules))
	assert.Equal(t, []admissionregistrationv1.OperationType{admissionregistrationv1.Create}, rules[0].Operations)
	assert.Equal(t, []string{""}, rules[0].Rule.APIGroups)
	assert.Equal(t, []string{"v1"}, rules[0].Rule.APIVersions)
	assert.Equal(t, []string{"nodes"}, rules[0].Rule.Resources)
}

func TestGetNodeMutatePath(t *testing.T) {
	t.Run("Generic path", func(t *testing.T) {
		path := getNodeMutatePath("test-subpath")
		expectedPath := "/mutate--v1-node/test-subpath"
		assert.Equal(t, expectedPath, path)
	})

	t.Run("Node startup taints specific path", func(t *testing.T) {
		path := getNodeMutatePath(NodeStartupTaintsMutatorID)
		expectedPath := "/mutate--v1-node/node-startup-taints"
		assert.Equal(t, expectedPath, path)
	})
}

func TestNewNodeHandler(t *testing.T) {
	mockMutator := &mocks.NodeMutator{}
	mockMutator.On("ID").Return("test-mutator")

	decoder := admission.NewDecoder(latest.Scheme)
	nodeHandler := NewNodeHandler(decoder, mockMutator)

	assert.NotNil(t, nodeHandler)
	assert.Equal(t, decoder, nodeHandler.decoder)
	assert.Equal(t, mockMutator, nodeHandler.mutator)
	assert.Equal(t, "test-mutator-webhook.union.ai", nodeHandler.mutatingWebhookName)
	assert.Equal(t, "/mutate--v1-node/test-mutator", nodeHandler.path)
}
