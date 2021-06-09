package webhook

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flytepropeller/pkg/webhook/config"

	"k8s.io/client-go/tools/clientcmd/api/latest"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/flyteorg/flytepropeller/pkg/webhook/mocks"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
)

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

	successMutator := &mocks.Mutator{}
	successMutator.OnID().Return("SucceedingMutator")
	successMutator.OnMutateMatch(mock.Anything, mock.Anything).Return(nil, false, nil)

	failedMutator := &mocks.Mutator{}
	failedMutator.OnID().Return("FailingMutator")
	failedMutator.OnMutateMatch(mock.Anything, mock.Anything).Return(nil, false, fmt.Errorf("failing mock"))

	t.Run("Required Mutator Succeeded", func(t *testing.T) {
		pm := &PodMutator{
			Mutators: []MutatorConfig{
				{
					Mutator:  successMutator,
					Required: true,
				},
			},
		}
		ctx := context.Background()
		_, changed, err := pm.Mutate(ctx, inputPod.DeepCopy())
		assert.NoError(t, err)
		assert.False(t, changed)
	})

	t.Run("Required Mutator Failed", func(t *testing.T) {
		pm := &PodMutator{
			Mutators: []MutatorConfig{
				{
					Mutator:  failedMutator,
					Required: true,
				},
			},
		}
		ctx := context.Background()
		_, _, err := pm.Mutate(ctx, inputPod.DeepCopy())
		assert.Error(t, err)
	})

	t.Run("Non-required Mutator Failed", func(t *testing.T) {
		pm := &PodMutator{
			Mutators: []MutatorConfig{
				{
					Mutator:  failedMutator,
					Required: false,
				},
			},
		}
		ctx := context.Background()
		_, _, err := pm.Mutate(ctx, inputPod.DeepCopy())
		assert.NoError(t, err)
	})
}

func Test_CreateMutationWebhookConfiguration(t *testing.T) {
	pm := NewPodMutator(&config.Config{
		CertDir:     "testdata",
		ServiceName: "my-service",
	}, promutils.NewTestScope())

	t.Run("Empty namespace", func(t *testing.T) {
		c, err := pm.CreateMutationWebhookConfiguration("")
		assert.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("With namespace", func(t *testing.T) {
		c, err := pm.CreateMutationWebhookConfiguration("my-namespace")
		assert.NoError(t, err)
		assert.NotNil(t, c)
	})
}

func Test_Handle(t *testing.T) {
	pm := NewPodMutator(&config.Config{
		CertDir:     "testdata",
		ServiceName: "my-service",
	}, promutils.NewTestScope())

	decoder, err := admission.NewDecoder(latest.Scheme)
	assert.NoError(t, err)
	assert.NoError(t, pm.InjectDecoder(decoder))

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
                "image": "bar:v2",
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
                "image": "bar:v1",
                "name": "bar"
            }
        ]
    }
}`),
			},
		},
	}

	resp := pm.Handle(context.Background(), req)
	assert.True(t, resp.Allowed)
}
