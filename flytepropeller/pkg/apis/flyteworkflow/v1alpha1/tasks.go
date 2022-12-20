package v1alpha1

import (
	"bytes"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/jsonpb"
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
	var buf bytes.Buffer
	if err := marshaler.Marshal(&buf, in.TaskTemplate); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (in *TaskSpec) UnmarshalJSON(b []byte) error {
	in.TaskTemplate = &core.TaskTemplate{}
	return jsonpb.Unmarshal(bytes.NewReader(b), in.TaskTemplate)
}
