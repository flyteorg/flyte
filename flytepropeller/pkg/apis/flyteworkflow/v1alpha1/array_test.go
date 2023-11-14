package v1alpha1

import (
	"testing"
)

func TestArrayNodeSpec_GetSubNodeSpec(t *testing.T) {
	nodeSpec := &NodeSpec{}
	arrayNodeSpec := ArrayNodeSpec{
		SubNodeSpec: nodeSpec,
	}

	if arrayNodeSpec.GetSubNodeSpec() != nodeSpec {
		t.Errorf("Expected nodeSpec, but got a different value")
	}
}

func TestArrayNodeSpec_GetParallelism(t *testing.T) {
	parallelism := uint32(5)
	arrayNodeSpec := ArrayNodeSpec{
		Parallelism: parallelism,
	}

	if arrayNodeSpec.GetParallelism() != parallelism {
		t.Errorf("Expected %d, but got %d", parallelism, arrayNodeSpec.GetParallelism())
	}
}

func TestArrayNodeSpec_GetMinSuccesses(t *testing.T) {
	minSuccesses := uint32(3)
	arrayNodeSpec := ArrayNodeSpec{
		MinSuccesses: &minSuccesses,
	}

	if *arrayNodeSpec.GetMinSuccesses() != minSuccesses {
		t.Errorf("Expected %d, but got %d", minSuccesses, *arrayNodeSpec.GetMinSuccesses())
	}
}

func TestArrayNodeSpec_GetMinSuccessRatio(t *testing.T) {
	minSuccessRatio := float32(0.8)
	arrayNodeSpec := ArrayNodeSpec{
		MinSuccessRatio: &minSuccessRatio,
	}

	if *arrayNodeSpec.GetMinSuccessRatio() != minSuccessRatio {
		t.Errorf("Expected %f, but got %f", minSuccessRatio, *arrayNodeSpec.GetMinSuccessRatio())
	}
}
