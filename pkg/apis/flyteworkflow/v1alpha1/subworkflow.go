package v1alpha1

type WorkflowNodeSpec struct {
	// Either one of the two
	LaunchPlanRefID *LaunchPlanRefID `json:"launchPlanRefId,omitempty"`
	// We currently want the SubWorkflow to be completely contained in the node. this is because
	// We use the node status to store the information of the execution.
	// Important Note: This may cause a bloat in case we use the same SubWorkflow in multiple nodes. The recommended
	// technique for that is to use launch plan refs. This is because we will end up executing the launch plan refs as
	// disparate executions in Flyte propeller. This is potentially better as it prevents us from hitting the storage limit
	// in etcd
	//+optional.
	// Workflow *WorkflowSpec `json:"workflow,omitempty"`
	SubWorkflowReference *WorkflowID `json:"subWorkflowRef,omitempty"`
}

func (in *WorkflowNodeSpec) GetLaunchPlanRefID() *LaunchPlanRefID {
	return in.LaunchPlanRefID
}

func (in *WorkflowNodeSpec) GetSubWorkflowRef() *WorkflowID {
	return in.SubWorkflowReference
}
