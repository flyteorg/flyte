package v1alpha1

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

type TaskSpec struct {
	*core.TaskTemplate
}

func (in *TaskSpec) TaskType() TaskType {
	return in.Type
}

func (in *TaskSpec) CoreTask() *core.TaskTemplate {
	return in.TaskTemplate
}

func (in *TaskSpec) DeepCopyInto(out *TaskSpec) {
	*out = *in
	// We do not manipulate the object, so its ok
	// Once we figure out the autogenerate story we can replace this
}

func (in *TaskSpec) MarshalJSON() ([]byte, error) {
	return utils.MarshalPbToBytes(in.TaskTemplate)
}

func (in *TaskSpec) UnmarshalJSON(b []byte) error {
	in.TaskTemplate = &core.TaskTemplate{}
	return utils.UnmarshalBytesToPb(b, in.TaskTemplate)
}
