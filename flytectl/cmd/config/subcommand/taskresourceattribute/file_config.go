package taskresourceattribute

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

// TaskResourceAttrFileConfig shadow Config for TaskResourceAttribute.
// The shadow Config is not using ProjectDomainAttribute/Workflowattribute directly inorder to simplify the inputs.
// As the same structure is being used for both ProjectDomainAttribute/Workflowattribute
type TaskResourceAttrFileConfig struct {
	Project  string `json:"project"`
	Domain   string `json:"domain"`
	Workflow string `json:"workflow,omitempty"`
	*admin.TaskResourceAttributes
}

// Decorate decorator over TaskResourceAttributes.
func (t TaskResourceAttrFileConfig) Decorate() *admin.MatchingAttributes {
	return &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_TaskResourceAttributes{
			TaskResourceAttributes: t.TaskResourceAttributes,
		},
	}
}

// UnDecorate to uncover TaskResourceAttributes.
func (t *TaskResourceAttrFileConfig) UnDecorate(matchingAttribute *admin.MatchingAttributes) {
	if matchingAttribute == nil {
		return
	}
	t.TaskResourceAttributes = matchingAttribute.GetTaskResourceAttributes()
}

// GetProject from the TaskResourceAttrFileConfig
func (t TaskResourceAttrFileConfig) GetProject() string {
	return t.Project
}

// GetDomain from the TaskResourceAttrFileConfig
func (t TaskResourceAttrFileConfig) GetDomain() string {
	return t.Domain
}

// GetWorkflow from the TaskResourceAttrFileConfig
func (t TaskResourceAttrFileConfig) GetWorkflow() string {
	return t.Workflow
}
