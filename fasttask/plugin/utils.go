package plugin

import (
	"encoding/json"
	"fmt"

	"k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"

	"github.com/unionai/flyte/fasttask/plugin/pb"
)

// isValidEnvironmentSpec validates the FastTaskEnvironmentSpec
func isValidEnvironmentSpec(executionEnvironmentID core.ExecutionEnvID, fastTaskEnvironmentSpec *pb.FastTaskEnvironmentSpec) error {
	if len(executionEnvironmentID.Name) == 0 {
		return fmt.Errorf("execution environment name is required")
	}

	if len(executionEnvironmentID.Version) == 0 {
		return fmt.Errorf("execution environment version is required")
	}

	if fastTaskEnvironmentSpec.GetBacklogLength() < 0 {
		return fmt.Errorf("backlog length must be greater than or equal to 0")
	}

	if fastTaskEnvironmentSpec.GetParallelism() <= 0 {
		return fmt.Errorf("parallelism must be greater than 0")
	}

	podTemplateSpec := &v1.PodTemplateSpec{}
	if err := json.Unmarshal(fastTaskEnvironmentSpec.GetPodTemplateSpec(), podTemplateSpec); err != nil {
		return fmt.Errorf("unable to unmarshal PodTemplateSpec [%v], Err: [%v]", fastTaskEnvironmentSpec.GetPodTemplateSpec(), err.Error())
	}

	if fastTaskEnvironmentSpec.GetReplicaCount() <= 0 {
		return fmt.Errorf("replica count must be greater than 0")
	}

	return nil
}

func isPodNotFoundErr(err error) bool {
	return k8serrors.IsNotFound(err) || k8serrors.IsGone(err)
}
