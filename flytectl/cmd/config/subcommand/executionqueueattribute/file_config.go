package executionqueueattribute

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// AttrFileConfig shadow Config for ExecutionQueueAttributes.
// The shadow Config is not using ProjectDomainAttribute/Workflowattribute directly inorder to simplify the inputs.
// As the same structure is being used for both ProjectDomainAttribute/Workflowattribute
type AttrFileConfig struct {
	Project  string `json:"project"`
	Domain   string `json:"domain"`
	Workflow string `json:"workflow,omitempty"`
	*admin.ExecutionQueueAttributes
}

// Decorate decorator over ExecutionQueueAttributes.
func (a AttrFileConfig) Decorate() *admin.MatchingAttributes {
	return &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
			ExecutionQueueAttributes: a.ExecutionQueueAttributes,
		},
	}
}

// UnDecorate to uncover ExecutionQueueAttributes.
func (a *AttrFileConfig) UnDecorate(matchingAttribute *admin.MatchingAttributes) {
	if matchingAttribute == nil {
		return
	}
	a.ExecutionQueueAttributes = matchingAttribute.GetExecutionQueueAttributes()
}

// GetProject from the AttrFileConfig
func (a AttrFileConfig) GetProject() string {
	return a.Project
}

// GetDomain from the AttrFileConfig
func (a AttrFileConfig) GetDomain() string {
	return a.Domain
}

// GetWorkflow from the AttrFileConfig
func (a AttrFileConfig) GetWorkflow() string {
	return a.Workflow
}
