package pluginoverride

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

// FileConfig shadow Config for PluginOverrides.
// The shadow Config is not using ProjectDomainAttribute/Workflowattribute directly inorder to simplify the inputs.
// As the same structure is being used for both ProjectDomainAttribute/Workflowattribute
type FileConfig struct {
	Project  string `json:"project"`
	Domain   string `json:"domain"`
	Workflow string `json:"workflow,omitempty"`
	*admin.PluginOverrides
}

// Decorate decorator over PluginOverrides.
func (t FileConfig) Decorate() *admin.MatchingAttributes {
	return &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_PluginOverrides{
			PluginOverrides: t.PluginOverrides,
		},
	}
}

// UnDecorate to uncover PluginOverrides.
func (t *FileConfig) UnDecorate(matchingAttribute *admin.MatchingAttributes) {
	if matchingAttribute == nil {
		return
	}
	t.PluginOverrides = matchingAttribute.GetPluginOverrides()
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
