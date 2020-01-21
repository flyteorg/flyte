package models

import "time"

type ResourcePriority int32

const (
	ResourcePriorityDomainLevel        ResourcePriority = 1
	ResourcePriorityProjectDomainLevel ResourcePriority = 10
	ResourcePriorityWorkflowLevel      ResourcePriority = 100
	ResourcePriorityLaunchPlanLevel    ResourcePriority = 1000
)

// Represents Flyte resources repository.
// In this model, the combination of (Project, Domain, Workflow, LaunchPlan, ResourceType) is unique
type Resource struct {
	ID           int64 `gorm:"AUTO_INCREMENT;column:id;primary_key"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `sql:"index"`
	Project      string     `gorm:"unique_index:resource_idx"`
	Domain       string     `gorm:"unique_index:resource_idx"`
	Workflow     string     `gorm:"unique_index:resource_idx"`
	LaunchPlan   string     `gorm:"unique_index:resource_idx"`
	ResourceType string     `gorm:"unique_index:resource_idx"`
	Priority     ResourcePriority
	// Serialized flyteidl.admin.MatchingAttributes.
	Attributes []byte
}
