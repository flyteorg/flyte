package compiler

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// This object is meant to satisfy github.com/flyteorg/flytepropeller/pkg/compiler/common.InterfaceProvider
// This file is pretty much copied from Admin, (sorry for the link, a real link made go mod import admin)
// github-dot-com/flyteorg/flyteadmin/blob/1acce744b8c7839ab77a0eb1ed922905af15baa5/pkg/workflowengine/impl/interface_provider.go
// but that implementation relies on the internal Admin Gorm model. We should consider deprecating that one in favor
// of this one as Admin already has a dependency on the Propeller compiler.
type LaunchPlanInterfaceProvider struct {
	expectedInputs  core.ParameterMap
	expectedOutputs core.VariableMap
	identifier      *core.Identifier
}

func (p *LaunchPlanInterfaceProvider) GetID() *core.Identifier {
	return p.identifier
}
func (p *LaunchPlanInterfaceProvider) GetExpectedInputs() *core.ParameterMap {
	return &p.expectedInputs

}
func (p *LaunchPlanInterfaceProvider) GetExpectedOutputs() *core.VariableMap {
	return &p.expectedOutputs
}

func NewLaunchPlanInterfaceProvider(launchPlan admin.LaunchPlan) *LaunchPlanInterfaceProvider {
	return &LaunchPlanInterfaceProvider{
		expectedInputs:  *launchPlan.Closure.ExpectedInputs,
		expectedOutputs: *launchPlan.Closure.ExpectedOutputs,
		identifier:      launchPlan.Id,
	}
}
