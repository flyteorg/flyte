package v1alpha1

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

type Identifier struct {
	*core.Identifier
}

func (in *Identifier) MarshalJSON() ([]byte, error) {
	return utils.MarshalPbToBytes(in.Identifier)
}

func (in *Identifier) UnmarshalJSON(b []byte) error {
	in.Identifier = &core.Identifier{}
	return utils.UnmarshalBytesToPb(b, in.Identifier)
}

func (in *Identifier) DeepCopyInto(out *Identifier) {
	*out = *in
}

type WorkflowExecutionIdentifier struct {
	*core.WorkflowExecutionIdentifier
}

func (in *WorkflowExecutionIdentifier) DeepCopyInto(out *WorkflowExecutionIdentifier) {
	*out = *in
}

type TaskExecutionIdentifier struct {
	*core.TaskExecutionIdentifier
}

func (in *TaskExecutionIdentifier) DeepCopyInto(out *TaskExecutionIdentifier) {
	*out = *in
}
