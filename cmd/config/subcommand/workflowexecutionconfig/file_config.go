package workflowexecutionconfig

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// FileConfig shadow Config for WorkflowExecutionConfig.
// The shadow Config is not using ProjectDomainAttribute/Workflowattribute directly inorder to simplify the inputs.
// As the same structure is being used for both ProjectDomainAttribute/Workflowattribute
type FileConfig struct {
	Project  string `json:"project"`
	Domain   string `json:"domain"`
	Workflow string `json:"workflow,omitempty"`
	*admin.WorkflowExecutionConfig
}

// Decorate decorator over WorkflowExecutionConfig.
func (t FileConfig) Decorate() *admin.MatchingAttributes {
	return &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
			WorkflowExecutionConfig: t.WorkflowExecutionConfig,
		},
	}
}

// UnDecorate to uncover WorkflowExecutionConfig.
func (t *FileConfig) UnDecorate(matchingAttribute *admin.MatchingAttributes) {
	if matchingAttribute == nil {
		return
	}
	t.WorkflowExecutionConfig = matchingAttribute.GetWorkflowExecutionConfig()
}

// GetProject from the WorkflowExecutionConfig
func (t FileConfig) GetProject() string {
	return t.Project
}

// GetDomain from the WorkflowExecutionConfig
func (t FileConfig) GetDomain() string {
	return t.Domain
}

// GetWorkflow from the WorkflowExecutionConfig
func (t FileConfig) GetWorkflow() string {
	return t.Workflow
}
