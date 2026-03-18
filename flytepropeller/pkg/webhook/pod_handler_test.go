package webhook

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd/api/latest"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/flyteorg/flyte/flytepropeller/pkg/secret"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/webhook/mocks"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// Use shared test constants from test_common.go

func TestPodHandler_Mutate(t *testing.T) {
	inputPod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "test",
				},
			},
		},
	}

	mockMutator := &mocks.PodMutator{}

	t.Run("Required Mutator Succeeded", func(t *testing.T) {
		mockMutator.On("Mutate", mock.Anything, inputPod).Return(inputPod, false, (*admission.Response)(nil)).Once()
		newPod, changed, admissionResp := mockMutator.Mutate(context.Background(), inputPod)
		assert.Nil(t, admissionResp)
		assert.False(t, changed)
		assert.Equal(t, inputPod, newPod)
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
		mockMutator.On("Mutate", mock.Anything, inputPod).Return(inputPod, false, errResp).Once()
		newPod, changed, admissionResp := mockMutator.Mutate(context.Background(), inputPod)
		assert.NotNil(t, admissionResp)
		assert.False(t, changed)
		assert.Equal(t, inputPod, newPod)
	})
}

func TestPodHandler_GetMutator(t *testing.T) {
	mockMutator := &mocks.PodMutator{}
	mockMutator.On("ID").Return("test-mutator")

	podHandler := NewPodHandler(admission.NewDecoder(latest.Scheme), mockMutator)

	assert.Equal(t, mockMutator, podHandler.GetMutator())
}

func TestPodHandler_GetPath(t *testing.T) {
	mockMutator := &mocks.PodMutator{}
	mockMutator.On("ID").Return("test-mutator")

	podHandler := NewPodHandler(admission.NewDecoder(latest.Scheme), mockMutator)

	expectedPath := "/mutate--v1-pod/test-mutator"
	assert.Equal(t, expectedPath, podHandler.GetPath())
}

func TestPodHandler_GetAdmissionHandler(t *testing.T) {
	mockMutator := &mocks.PodMutator{}
	mockMutator.On("ID").Return("test-mutator")

	podHandler := NewPodHandler(admission.NewDecoder(latest.Scheme), mockMutator)

	handler := podHandler.GetAdmissionHandler()
	assert.Equal(t, podHandler, handler)
}

func TestInitializePodHandlers(t *testing.T) {
	ctx := context.Background()

	t.Run("Secrets only", func(t *testing.T) {
		cfg := &config.Config{
			EmbeddedSecretManagerConfig: secretManagerConfig,
		}

		handlers, _, err := initializePodHandlers(ctx, cfg, latest.Scheme, testNamespace, promutils.NewTestScope())
		assert.NoError(t, err)
		assert.Equal(t, 1, len(handlers))

		// Verify it's a pod handler with secrets mutator
		podHandler, ok := handlers[0].(*PodHandler)
		assert.True(t, ok)
		assert.Equal(t, secret.SecretsID, podHandler.GetMutator().ID())
	})

	t.Run("With image builder", func(t *testing.T) {
		cfg := &config.Config{
			EmbeddedSecretManagerConfig: secretManagerConfig,
			ImageBuilderConfig:          testImageBuilderConfig,
		}

		handlers, _, err := initializePodHandlers(ctx, cfg, latest.Scheme, testNamespace, promutils.NewTestScope())
		assert.NoError(t, err)
		assert.Equal(t, 2, len(handlers))

		// Verify first handler is secrets
		podHandler1, ok := handlers[0].(*PodHandler)
		assert.True(t, ok)
		assert.Equal(t, secret.SecretsID, podHandler1.GetMutator().ID())

		// Verify second handler is image builder
		podHandler2, ok := handlers[1].(*PodHandler)
		assert.True(t, ok)
		assert.Equal(t, ImageBuilderV1ID, podHandler2.GetMutator().ID())
	})
}
