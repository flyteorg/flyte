package yunikorn

import (
	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
)

const (
	// Pod lebel
	Yunikorn             = "yunikorn"
	AppID                = "yunikorn.apache.org/app-id"
	TaskGroupNameKey     = "yunikorn.apache.org/task-group-name"
	TaskGroupsKey        = "yunikorn.apache.org/task-groups"
	TaskGroupPrarameters = "yunikorn.apache.org/schedulingPolicyParameters"
)

type Plugin struct {
	Parameters  string
}

func NewPlugin(parameters string) *Plugin {
	return &Plugin{
		Parameters:  parameters,
	}
}

func (p *Plugin) Process(app interface{}) error {
	switch v := app.(type) {
	case *rayv1.RayJob:
		return ProcessRay(p.Parameters, v)
	default:
		return nil
	}
}
