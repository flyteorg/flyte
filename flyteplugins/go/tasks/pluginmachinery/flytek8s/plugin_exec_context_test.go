package flytek8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_NodeExecutionK8sReader(t *testing.T) {
	execID := "abc123"
	nodeID := "n0"
	typeMeta := metav1.TypeMeta{
		Kind:       "Pod",
		APIVersion: "v1",
	}
	pod1 := v1.Pod{
		TypeMeta: typeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "a",
			Namespace: namespace,
			Labels: map[string]string{
				"some-label":   "bar",
				"execution-id": execID,
				"node-id":      nodeID,
			},
		},
	}
	pod2 := v1.Pod{
		TypeMeta: typeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "b",
			Namespace: namespace,
			Labels: map[string]string{
				"execution-id": execID,
				"node-id":      nodeID,
			},
		},
	}
	pod3 := v1.Pod{
		TypeMeta: typeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "c",
			Namespace: "wrong",
			Labels: map[string]string{
				"execution-id": execID,
				"node-id":      nodeID,
			},
		},
	}
	pod4 := v1.Pod{
		TypeMeta: typeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "d",
			Namespace: namespace,
			Labels: map[string]string{
				"execution-id": "wrong",
				"node-id":      nodeID,
			},
		},
	}
	pod5 := v1.Pod{
		TypeMeta: typeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "e",
			Namespace: namespace,
			Labels: map[string]string{
				"execution-id": execID,
				"node-id":      "wrong",
			},
		},
	}
	pods := []runtime.Object{&pod1, &pod2, &pod3, &pod4, &pod5}
	nodeExecReader := NodeExecutionK8sReader{
		namespace:   namespace,
		executionID: execID,
		nodeID:      nodeID,
		client:      fake.NewFakeClient(pods...),
	}
	ctx := context.TODO()

	t.Run("get", func(t *testing.T) {
		p := v1.Pod{}

		err := nodeExecReader.Get(ctx, client.ObjectKeyFromObject(&pod1), &p)

		assert.NoError(t, err)
		assert.Equal(t, pod1, p)
	})

	t.Run("get-not-found", func(t *testing.T) {
		p := v1.Pod{}

		for _, input := range []*v1.Pod{&pod3, &pod4, &pod5} {
			err := nodeExecReader.Get(ctx, client.ObjectKeyFromObject(input), &p)

			assert.True(t, errors.IsNotFound(err))
		}
	})

	t.Run("list", func(t *testing.T) {
		p := &v1.PodList{}
		expected := &v1.PodList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodList",
				APIVersion: "v1",
			},
			Items: []v1.Pod{pod1, pod2},
		}

		err := nodeExecReader.List(ctx, p)

		assert.NoError(t, err)
		assert.Equal(t, expected, p)
	})
}
