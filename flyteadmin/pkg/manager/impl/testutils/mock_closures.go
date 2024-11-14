// Provides sample closures for use in tests.
package testutils

import (
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

var MockCreatedAtValue = time.Date(2018, time.February, 17, 00, 00, 00, 00, time.UTC).UTC()
var MockCreatedAtProto, _ = ptypes.TimestampProto(MockCreatedAtValue)

func GetTaskClosure() *admin.TaskClosure {
	return &admin.TaskClosure{
		CompiledTask: &core.CompiledTask{
			Template: GetValidTaskRequest().GetSpec().GetTemplate(),
		},
		CreatedAt: MockCreatedAtProto,
	}
}

func GetTaskClosureBytes() []byte {
	var taskClosureBytes, _ = proto.Marshal(GetTaskClosure())
	return taskClosureBytes
}

func GetWorkflowClosure() *admin.WorkflowClosure {
	return &admin.WorkflowClosure{
		CompiledWorkflow: &core.CompiledWorkflowClosure{
			Primary: &core.CompiledWorkflow{
				Template: GetWorkflowRequest().GetSpec().GetTemplate(),
			},
			Tasks: []*core.CompiledTask{
				{
					Template: GetValidTaskRequest().GetSpec().GetTemplate(),
				},
			},
		},
		CreatedAt: MockCreatedAtProto,
	}
}

func GetWorkflowClosureBytes() []byte {
	// WorkflowClosure
	var workflowClosureBytes, _ = proto.Marshal(GetWorkflowClosure())
	return workflowClosureBytes
}

func MakeStringLiteral(value string) *core.Literal {
	p := &core.Primitive{
		Value: &core.Primitive_StringValue{
			StringValue: value,
		},
	}
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: p,
				},
			},
		},
	}
}
