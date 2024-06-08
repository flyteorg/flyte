package v1alpha1

import (
	"bytes"

	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

type Identifier struct {
	*core.Identifier
}

func (in *Identifier) UnmarshalJSON(b []byte) error {
	in.Identifier = &core.Identifier{}
	return jsonpb.Unmarshal(bytes.NewReader(b), in.Identifier)
}

func (in *Identifier) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	if err := marshaler.Marshal(&buf, in.Identifier); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (in *Identifier) DeepCopyInto(out *Identifier) {
	out.Identifier = proto.Clone(in.Identifier).(*core.Identifier)
}

type WorkflowExecutionIdentifier struct {
	*core.WorkflowExecutionIdentifier
}

func (in *WorkflowExecutionIdentifier) DeepCopyInto(out *WorkflowExecutionIdentifier) {
	out.WorkflowExecutionIdentifier = proto.Clone(in.WorkflowExecutionIdentifier).(*core.WorkflowExecutionIdentifier)
}

type TaskExecutionIdentifier struct {
	*core.TaskExecutionIdentifier
}

func (in *TaskExecutionIdentifier) DeepCopyInto(out *TaskExecutionIdentifier) {
	out.TaskExecutionIdentifier = proto.Clone(in.TaskExecutionIdentifier).(*core.TaskExecutionIdentifier)
}
