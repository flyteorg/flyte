package plugin

import (
	"encoding/json"
	"fmt"

	"k8s.io/api/core/v1"

	"github.com/unionai/flyte/fasttask/plugin/pb"
)

// isValidEnvironmentSpec validates the FastTaskEnvironmentSpec
func isValidEnvironmentSpec(fastTaskEnvironmentSpec *pb.FastTaskEnvironmentSpec) error {
	if fastTaskEnvironmentSpec.GetBacklogLength() < 0 {
		return fmt.Errorf("backlog length must be greater than or equal to 0")
	}

	if fastTaskEnvironmentSpec.GetParallelism() <= 0 {
		return fmt.Errorf("parallelism must be greater than 0")
	}

	podTemplateSpec := &v1.PodTemplateSpec{}
	if err := json.Unmarshal(fastTaskEnvironmentSpec.PodTemplateSpec, podTemplateSpec); err != nil {
		return fmt.Errorf("unable to unmarshal PodTemplateSpec [%v], Err: [%v]", fastTaskEnvironmentSpec.PodTemplateSpec, err.Error())
	}

	if fastTaskEnvironmentSpec.GetReplicaCount() <= 0 {
		return fmt.Errorf("replica count must be greater than 0")
	}

	return nil
}
