package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMarshalPatch(t *testing.T) {
	patch := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec: appsv1.DeploymentSpec{},
	}
	t.Run("NoFilter", func(t *testing.T) {
		y, err := MarshalPatch(&patch, false, false)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(
			t,
			string(y),
			`apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  name: foo
  namespace: bar
spec:
  selector: null
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
    spec:
      containers: null
status: {}
`,
			"YAML strings should match",
		)
	})
	t.Run("FilterNil", func(t *testing.T) {
		y, err := MarshalPatch(&patch, true, false)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(
			t,
			string(y),
			`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  namespace: bar
spec:
  strategy: {}
  template:
    metadata: {}
    spec: {}
status: {}
`,
			"YAML strings should match",
		)
	})
	t.Run("FilterEmpty", func(t *testing.T) {
		y, err := MarshalPatch(&patch, false, true)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(
			t,
			string(y),
			`apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  name: foo
  namespace: bar
spec:
  selector: null
  template:
    metadata:
      creationTimestamp: null
    spec:
      containers: null
`,
			"YAML strings should match",
		)
	})
	t.Run("FilterAll", func(t *testing.T) {
		y, err := MarshalPatch(&patch, true, true)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(
			t,
			string(y),
			`apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  namespace: bar
`,
			"YAML strings should match",
		)
	})
}
