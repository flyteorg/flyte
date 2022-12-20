package executionclusterlabel

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// FileConfig shadow Config for ExecutionClusterLabel.
// The shadow Config is not using ProjectDomainAttribute/Workflowattribute directly inorder to simplify the inputs.
// As the same structure is being used for both ProjectDomainAttribute/Workflowattribute
type FileConfig struct {
	Project  string `json:"project"`
	Domain   string `json:"domain"`
	Workflow string `json:"workflow,omitempty"`
	*admin.ExecutionClusterLabel
}

// Decorate decorator over ExecutionClusterLabel.
func (t FileConfig) Decorate() *admin.MatchingAttributes {
	return &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_ExecutionClusterLabel{
			ExecutionClusterLabel: t.ExecutionClusterLabel,
		},
	}
}

// UnDecorate to uncover ExecutionClusterLabel.
func (t *FileConfig) UnDecorate(matchingAttribute *admin.MatchingAttributes) {
	if matchingAttribute == nil {
		return
	}
	t.ExecutionClusterLabel = matchingAttribute.GetExecutionClusterLabel()
}

// GetProject from the FileConfig
func (t FileConfig) GetProject() string {
	return t.Project
}

// GetDomain from the FileConfig
func (t FileConfig) GetDomain() string {
	return t.Domain
}

// GetWorkflow from the FileConfig
func (t FileConfig) GetWorkflow() string {
	return t.Workflow
}
