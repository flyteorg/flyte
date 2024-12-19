package v1alpha1

import "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"

type ArrayNodeSpec struct {
	SubNodeSpec     *NodeSpec
	Parallelism     *uint32
	MinSuccesses    *uint32
	MinSuccessRatio *float32
	ExecutionMode   core.ArrayNode_ExecutionMode
	DataMode        core.ArrayNode_DataMode
}

func (a *ArrayNodeSpec) GetSubNodeSpec() *NodeSpec {
	return a.SubNodeSpec
}

func (a *ArrayNodeSpec) GetParallelism() *uint32 {
	return a.Parallelism
}

func (a *ArrayNodeSpec) GetMinSuccesses() *uint32 {
	return a.MinSuccesses
}

func (a *ArrayNodeSpec) GetMinSuccessRatio() *float32 {
	return a.MinSuccessRatio
}

func (a *ArrayNodeSpec) GetExecutionMode() core.ArrayNode_ExecutionMode {
	return a.ExecutionMode
}

func (a *ArrayNodeSpec) GetDataMode() core.ArrayNode_DataMode {
	return a.DataMode
}
