package shardstrategy

import (
	"fmt"

	"github.com/flyteorg/flytepropeller/pkg/utils"

	v1 "k8s.io/api/core/v1"
)

// EnvironmentShardStrategy assigns either project or domain identifers to individual
// FlytePropeller instances to determine FlyteWorkflow processing responsibility.
type EnvironmentShardStrategy struct {
	EnvType     environmentType
	PerShardIDs [][]string
}

type environmentType int

const (
	Project environmentType = iota
	Domain
)

func (e environmentType) String() string {
	return [...]string{"project", "domain"}[e]
}

func (e *EnvironmentShardStrategy) GetPodCount() int {
	return len(e.PerShardIDs)
}

func (e *EnvironmentShardStrategy) HashCode() (uint32, error) {
	return computeHashCode(e)
}

func (e *EnvironmentShardStrategy) UpdatePodSpec(pod *v1.PodSpec, containerName string, podIndex int) error {
	container, err := utils.GetContainer(pod, containerName)
	if err != nil {
		return err
	}

	if podIndex < 0 || podIndex >= e.GetPodCount() {
		return fmt.Errorf("invalid podIndex '%d' out of range [0,%d)", podIndex, e.GetPodCount())
	}

	if len(e.PerShardIDs[podIndex]) == 1 && e.PerShardIDs[podIndex][0] == "*" {
		for i, shardIDs := range e.PerShardIDs {
			if i != podIndex {
				for _, id := range shardIDs {
					container.Args = append(container.Args, fmt.Sprintf("--propeller.exclude-%s-label", e.EnvType), id)
				}
			}
		}
	} else {
		for _, id := range e.PerShardIDs[podIndex] {
			container.Args = append(container.Args, fmt.Sprintf("--propeller.include-%s-label", e.EnvType), id)
		}
	}

	return nil
}
