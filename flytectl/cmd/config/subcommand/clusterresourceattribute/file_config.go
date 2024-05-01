package clusterresourceattribute

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

// AttrFileConfig shadow Config for ClusterResourceAttributes.
// The shadow Config is not using ProjectDomainAttribute/Workflowattribute directly inorder to simplify the inputs.
// As the same structure is being used for both ProjectDomainAttribute/Workflowattribute
type AttrFileConfig struct {
	Project  string `json:"project"`
	Domain   string `json:"domain"`
	Workflow string `json:"workflow,omitempty"`
	*admin.ClusterResourceAttributes
}

// Decorate decorator over ClusterResourceAttributes.
func (c AttrFileConfig) Decorate() *admin.MatchingAttributes {
	return &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_ClusterResourceAttributes{
			ClusterResourceAttributes: c.ClusterResourceAttributes,
		},
	}
}

// UnDecorate to uncover ClusterResourceAttributes.
func (c *AttrFileConfig) UnDecorate(matchingAttribute *admin.MatchingAttributes) {
	if matchingAttribute == nil {
		return
	}
	c.ClusterResourceAttributes = matchingAttribute.GetClusterResourceAttributes()
}

// GetProject from the AttrFileConfig
func (c AttrFileConfig) GetProject() string {
	return c.Project
}

// GetDomain from the AttrFileConfig
func (c AttrFileConfig) GetDomain() string {
	return c.Domain
}

// GetWorkflow from the AttrFileConfig
func (c AttrFileConfig) GetWorkflow() string {
	return c.Workflow
}
