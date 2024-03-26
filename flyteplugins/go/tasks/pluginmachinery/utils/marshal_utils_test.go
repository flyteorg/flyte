package utils

import (
	"encoding/json"
	"testing"

	"github.com/go-test/deep"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestUnmarshalStructToObj(t *testing.T) {
	t.Run("no nil structs allowed", func(t *testing.T) {
		var podSpec v1.PodSpec
		err := UnmarshalStructToObj(nil, &podSpec)
		assert.EqualError(t, err, "nil Struct Object passed")
	})
	podSpec := v1.PodSpec{
		Containers: []v1.Container{
			{
				Name: "a container",
			},
			{
				Name: "another container",
			},
		},
	}

	b, err := json.Marshal(podSpec)
	if err != nil {
		t.Fatal(err)
	}

	structObj := &structpb.Struct{}
	if err := json.Unmarshal(b, structObj); err != nil {
		t.Fatal(err)
	}

	t.Run("no nil pointers as obj allowed", func(t *testing.T) {
		var nilPodspec *v1.PodSpec
		err := UnmarshalStructToObj(structObj, nilPodspec)
		assert.EqualError(t, err, "json: Unmarshal(nil *v1.PodSpec)")
	})

	t.Run("happy case", func(t *testing.T) {
		var podSpecObj v1.PodSpec
		err := UnmarshalStructToObj(structObj, &podSpecObj)
		assert.NoError(t, err)
		if diff := deep.Equal(podSpecObj, podSpec); diff != nil {
			t.Errorf("UnmarshalStructToObj() got = %v, want %v, diff: %v", podSpecObj, podSpec, diff)
		}
	})
}
