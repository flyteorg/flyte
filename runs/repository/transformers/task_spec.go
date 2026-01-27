package transformers

import (
	"google.golang.org/protobuf/proto"

	flyteWorkflow "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// ToTaskSpec converts TaskSpec model to proto TaskSpec
func ToTaskSpec(specModel *models.TaskSpec) (*flyteWorkflow.TaskSpec, error) {
	spec := &flyteWorkflow.TaskSpec{}
	if err := proto.Unmarshal(specModel.Spec, spec); err != nil {
		return nil, err
	}
	return spec, nil
}

// ToTraceSpec converts TaskSpec model to proto TraceSpec
func ToTraceSpec(specModel *models.TaskSpec) (*flyteWorkflow.TraceSpec, error) {
	spec := &flyteWorkflow.TraceSpec{}
	if err := proto.Unmarshal(specModel.Spec, spec); err != nil {
		return nil, err
	}
	return spec, nil
}
