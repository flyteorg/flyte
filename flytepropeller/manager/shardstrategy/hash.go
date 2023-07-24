package shardstrategy

import (
	"fmt"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/utils"

	v1 "k8s.io/api/core/v1"
)

// HashShardStrategy evenly assigns disjoint keyspace responsibilities over a collection of pods.
// All FlyteWorkflows are assigned a shard-key using a hash of their executionID and are then
// processed by the FlytePropeller instance responsible for that keyspace range.
type HashShardStrategy struct {
	ShardCount int
}

func (h *HashShardStrategy) GetPodCount() int {
	return h.ShardCount
}

func (h *HashShardStrategy) HashCode() (uint32, error) {
	return computeHashCode(h)
}

func (h *HashShardStrategy) UpdatePodSpec(pod *v1.PodSpec, containerName string, podIndex int) error {
	container, err := utils.GetContainer(pod, containerName)
	if err != nil {
		return err
	}

	if podIndex < 0 || podIndex >= h.GetPodCount() {
		return fmt.Errorf("invalid podIndex '%d' out of range [0,%d)", podIndex, h.GetPodCount())
	}

	startKey, endKey := ComputeKeyRange(v1alpha1.ShardKeyspaceSize, h.GetPodCount(), podIndex)
	for i := startKey; i < endKey; i++ {
		container.Args = append(container.Args, "--propeller.include-shard-key-label", fmt.Sprintf("%d", i))
	}

	return nil
}

// ComputeKeyRange computes a [startKey, endKey) pair denoting the key responsibilities for the
// provided pod index given the keyspaceSize and podCount parameters.
func ComputeKeyRange(keyspaceSize, podCount, podIndex int) (int, int) {
	keysPerPod := keyspaceSize / podCount
	keyRemainder := keyspaceSize - (podCount * keysPerPod)

	return computeStartKey(keysPerPod, keyRemainder, podIndex), computeStartKey(keysPerPod, keyRemainder, podIndex+1)
}

func computeStartKey(keysPerPod, keysRemainder, podIndex int) int {
	return (intMin(podIndex, keysRemainder) * (keysPerPod + 1)) + (intMax(0, podIndex-keysRemainder) * keysPerPod)
}

func intMin(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func intMax(a, b int) int {
	if a > b {
		return a
	}

	return b
}
